import requests
from requests.auth import HTTPBasicAuth

churl = 'https://rc1a-psxonv91j7bbal7v.mdb.yandexcloud.net:8443'
chuser = 'events'
chpass = 'password'
chinsert = 'INSERT INTO events.events (ts, gender, age, path, browser, os) FORMAT JSONEachRow'

def main(event, context):
    payload = ''
    count = 0
    for msg in event['messages']:
        body = msg['details']['message']['body']
        payload = payload + body + '\n'
        count = count + 1
    
    r = requests.post(churl + '/?query=' + chinsert, data = payload, auth=HTTPBasicAuth(chuser, chpass), verify=False)
    print("Pushed " + str(count) + " messages to clickhouse")
