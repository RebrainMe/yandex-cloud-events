spec:
  containers:
    - name: api
      image: vozerov/events-api:v3
      command:
        - /app/app
      args:
        - -kafka=${kafka_uri}
        - -enable-kafka
        - -amqp=${amqp_uri}
        - -enable-amqp
        - -sqs-uri=https://message-queue.api.cloud.yandex.net/b1gv67ihgfu3bpt24o0q/dj6000000000m4t607bu/load
        - -sqs-id=veXDMRkg4LqWsmMm1Kes
        - -sqs-secret=H09WgMAqp-VyxIxPFHnd_Py2azsw5U36sv6INs-K
        - -enable-sqs
      securityContext:
        privileged: false
      tty: false
      stdin: false
      restartPolicy: Always
