#!/usr/bin/python3

import random,sys

count = 0
if len(sys.argv) == 2:
  count = int(sys.argv[1])

print("Got count: " + str(count))
endpoint = ['sqs', 'kafka', 'kafka', 'kafka']

gender = ['male', 'female']
age = 0.00

path = ['/index', '/buy', '/kids', '/adult']

browser = ['chrome', 'safari', 'edge', 'netscape']
os = ['windows', 'osx', 'linux', 'ios', 'android']

ts_start = 1583275914
ts_end = 1583707917

with open('data.txt', 'w') as f:
  for x in range(count):
    print('POST||/post/' + random.choice(endpoint) + '||json||{"ts": ' + str(random.randrange(ts_start, ts_end)) + ', "gender":"' + random.choice(gender) + '", "age":' + str(random.randrange(18, 60)) + ', "path": "' + random.choice(path) + '", "browser": "' + random.choice(browser) + '", "os": "' + random.choice(os)+ '"}', file=f)

