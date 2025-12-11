#!/usr/bin/env python3
"""
Deploy Grafana Dashboard to Fly.io Grafana

This script programmatically deploys the Grafana dashboard using Grafana's HTTP API.

Usage:
    export GRAFANA_URL="https://your-grafana-url"
    export GRAFANA_API_KEY="your-api-key"
    python3 scripts/deploy-grafana-dashboard.py

Or with command line arguments:
    python3 scripts/deploy-grafana-dashboard.py --url https://your-grafana-url --key your-api-key
"""

import json
import os
import sys
import argparse
from pathlib import Path
from urllib.request import Request, urlopen
from urllib.error import HTTPError, URLError


class Colors:
    """ANSI color codes for terminal output"""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color


def print_success(msg):
    print(f"{Colors.GREEN}{msg}{Colors.NC}")


def print_error(msg):
    print(f"{Colors.RED}{msg}{Colors.NC}")


def print_warning(msg):
    print(f"{Colors.YELLOW}{msg}{Colors.NC}")


def print_info(msg):
    print(f"{Colors.BLUE}{msg}{Colors.NC}")


def load_dashboard(dashboard_path: Path) -> dict:
    """Load dashboard JSON from file"""
    if not dashboard_path.exists():
        raise FileNotFoundError(f"Dashboard file not found: {dashboard_path}")
    
    with open(dashboard_path, 'r') as f:
        return json.load(f)


def deploy_dashboard(grafana_url: str, api_key: str, dashboard: dict) -> dict:
    """Deploy dashboard to Grafana using HTTP API"""
    
    # Prepare payload for Grafana API
    payload = {
        "dashboard": dashboard.get("dashboard", dashboard),
        "overwrite": True,
        "message": "Deployed via automation script"
    }
    
    # Prepare request
    url = f"{grafana_url}/api/dashboards/db"
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }
    
    data = json.dumps(payload).encode('utf-8')
    request = Request(url, data=data, headers=headers, method='POST')
    
    try:
        with urlopen(request) as response:
            result = json.loads(response.read().decode('utf-8'))
            return {
                "success": True,
                "status_code": response.status,
                "data": result
            }
    except HTTPError as e:
        error_body = e.read().decode('utf-8')
        try:
            error_json = json.loads(error_body)
        except json.JSONDecodeError:
            error_json = {"message": error_body}
        
        return {
            "success": False,
            "status_code": e.code,
            "error": error_json
        }
    except URLError as e:
        return {
            "success": False,
            "status_code": None,
            "error": {"message": str(e.reason)}
        }


def show_manual_instructions(dashboard_path: Path):
    """Show instructions for manual dashboard import"""
    print()
    print_warning("To deploy to Fly.io Grafana:")
    print("1. Get your Grafana URL from: https://fly.io/dashboard")
    print("2. Navigate to your organization's Grafana instance")
    print("3. Go to: Settings → API Keys → New API Key")
    print("   - Name: 'Dashboard Deployment'")
    print("   - Role: 'Editor'")
    print("   - Time to live: Set as needed")
    print()
    print("4. Set environment variables:")
    print("   export GRAFANA_URL='https://your-grafana-url'")
    print("   export GRAFANA_API_KEY='your-api-key'")
    print()
    print("5. Run this script again:")
    print(f"   python3 {__file__}")
    print()
    print_info("Alternative: Manual import")
    print(f"Upload {dashboard_path} manually via Grafana UI:")
    print("  Dashboard → Import → Upload JSON file")
    print()


def main():
    parser = argparse.ArgumentParser(
        description='Deploy Grafana dashboard to Fly.io Grafana'
    )
    parser.add_argument(
        '--url',
        help='Grafana URL (or set GRAFANA_URL env var)',
        default=os.getenv('GRAFANA_URL')
    )
    parser.add_argument(
        '--key',
        help='Grafana API key (or set GRAFANA_API_KEY env var)',
        default=os.getenv('GRAFANA_API_KEY')
    )
    parser.add_argument(
        '--dashboard',
        help='Path to dashboard JSON file',
        type=Path,
        default=None
    )
    
    args = parser.parse_args()
    
    print("=" * 45)
    print("Grafana Dashboard Deployment Script")
    print("=" * 45)
    print()
    
    # Determine dashboard path
    if args.dashboard:
        dashboard_path = args.dashboard
    else:
        script_dir = Path(__file__).parent
        project_root = script_dir.parent
        dashboard_path = project_root / "docs" / "grafana-dashboard.json"
    
    # Check if dashboard file exists
    if not dashboard_path.exists():
        print_error(f"Error: Dashboard file not found: {dashboard_path}")
        sys.exit(1)
    
    # Check for required parameters
    if not args.url:
        show_manual_instructions(dashboard_path)
        sys.exit(1)
    
    if not args.key:
        print_error("Error: GRAFANA_API_KEY is required")
        print("Create an API key in Grafana: Settings → API Keys → New API Key")
        print("Then: export GRAFANA_API_KEY='your-api-key'")
        sys.exit(1)
    
    print(f"Grafana URL: {args.url}")
    print(f"Dashboard file: {dashboard_path}")
    print()
    
    # Load dashboard
    try:
        dashboard = load_dashboard(dashboard_path)
    except Exception as e:
        print_error(f"Error loading dashboard: {e}")
        sys.exit(1)
    
    # Deploy dashboard
    print("Deploying dashboard to Grafana...")
    result = deploy_dashboard(args.url, args.key, dashboard)
    
    if result["success"]:
        print_success("✓ Dashboard deployed successfully!")
        print()
        
        # Extract dashboard info from response
        data = result.get("data", {})
        dashboard_url = data.get("url", "")
        dashboard_uid = data.get("uid", "")
        dashboard_id = data.get("id", "")
        
        if dashboard_url:
            print(f"Dashboard URL: {args.url}{dashboard_url}")
        if dashboard_uid:
            print(f"Dashboard UID: {dashboard_uid}")
        if dashboard_id:
            print(f"Dashboard ID: {dashboard_id}")
        
        print()
        print_success("Dashboard is now live!")
        sys.exit(0)
    else:
        print_error("✗ Failed to deploy dashboard")
        print(f"HTTP Status: {result.get('status_code', 'N/A')}")
        print("Error:")
        print(json.dumps(result.get("error", {}), indent=2))
        sys.exit(1)


if __name__ == "__main__":
    main()
