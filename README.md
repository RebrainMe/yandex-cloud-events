# Events API Yandex Cloud

Collection of tools that were used on rebrainme's webinar with yandex cloud.

# Tools!

- **terraform** - terraform files for container registry, service accounts, instances, instance group, load balancer, cloudflare's dns
- **ansible** - ansible playbooks for docker, kafka, grafana, prometheus
- **app** - small application server which receives json events through http and push them to kafka (should be rewritten with fasthttp and some optimizations).
- **grafana** - dashboards used at webinar
- **load** - yandex tank configration for load testing

# Tested versions
- terraform:
```sh
$ terraform version
Terraform v0.12.18
```

- ansible:
```sh
$ ansible --version
ansible 2.9.4
```

- golang: 1.13.8 from docker golang:latest

- yandex tank:
```sh
# docker run -v /home/cloud-user/load/:/var/loadtest/ --net host -it direvius/yandex-tank --version
No handlers could be found for logger "netort.resource"
YandexTank/1.12.1
```

# Terraform
To run terraform you should create **private.auto.tfvars** file with the following content:
```ini
yc_token = "YANDEX CLOUD OAUTH TOKEN"
yc_cloud_id = "YANDEX CLOUD ID"
yc_folder_id = "YANDEX CLOUD FOLDER ID"
cf_email = "CLOUDFLARE EMAIL"
cf_token = "CLOUDFLARE TOKEN"
cf_zone_id = "CLOUDFLARE ZONE ID"
```

Then just run:
```sh
$ terraform apply
```

# Ansible
In fact, ansible is not used directly - it is run by terraform to provision kafka, build and monitoring hosts. Anyway, you have to download all the roles from anible galaxy:
```sh
$ ansible-galaxy install -r requirements.yml
```

Also, you can use provided .ansible.cfg file.

# Load
First of all signup to [yandex overload](http://overload.yandex.net) if you want to get some visualisation on the web. After you're in - obtain your api key and write it down to **token.txt** file.

After that you can easily run load test using docker:
```sh
docker run -v $(pwd):/var/loadtest/ --net host -it direvius/yandex-tank -c load.yml ammo.txt
```

**ammo.txt** can be generated with ammo generators - [Official doc](https://yandextank.readthedocs.io/en/latest/ammo_generators.html)

# Grafana
Just import dashboard to your grafana instance and select proper prometheus data source - thats it!

# Application
Use docker to build the container:
```sh
$ docker build -t app .
```

Then you can run it with:
```sh
$ docker run -d --name api -p 8080:8080 api
```

Some flags are used by the app:
- **-addr** - address to listen to. By default - :8080
- **-kafka** - separated list of kafka brokers to use (i.e. 10.0.0.1:9092,10.0.0.2:9092). By default - 127.0.0.1:9092

License
----
MIT

Copyright (c) 2020 Vasiliy I Ozerov, Rebrain

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

