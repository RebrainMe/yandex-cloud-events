spec:
  containers:
    - name: streaming
      image: vozerov/events-streaming:v1
      command:
        - /app/streaming
      args:
        - -kafka-brokers=${kafka_uri}
        - -kafka-topic=loader
        - -ch-dsn=tcp://${clickhouse.host[0].fqdn}:9440?username=events&password=password&database=events&secure=true&skip_verify=true
      securityContext:
        privileged: false
      tty: false
      stdin: false
      restartPolicy: Always
