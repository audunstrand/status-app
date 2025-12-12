FROM grafana/alloy:latest

COPY grafana-agent/config.alloy /etc/alloy/config.alloy

ENTRYPOINT ["/bin/alloy"]
CMD ["run", "/etc/alloy/config.alloy", "--server.http.listen-addr=0.0.0.0:8080"]
