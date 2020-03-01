resource "yandex_compute_instance" "zookeeper" {
  count = 3
  name = "zookeeper-${count.index+1}"
  zone = "ru-central1-a"
  hostname = "zookeeper-${count.index+1}"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 4
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size = 15
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.internal-a.id
    nat       = true
  }

  metadata = {
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
  }

  # remote-exec will wait for ssh up and running, after that local-exec will come into play
  # XXX By default requires SSH agent to be running
  provisioner "remote-exec" {
    inline = ["# Connected!"]
    connection {
      host = self.network_interface.0.nat_ip_address
      user = "ubuntu"
    }
  }
}

resource "null_resource" "zookeeper_deploy" {
  triggers = {
    zookeeper_nodes_ids = "${join(",", yandex_compute_instance.zookeeper[*].id)}"
  }

  # Generate Ansible inventory file
  provisioner "local-exec" {
    command = <<-EOA
    echo "${templatefile("${path.module}/templates/ansible_inventory_zookeeper.yml.tpl", { zookeeper_nodes = yandex_compute_instance.zookeeper[*] })}" > ${var.ansible_workdir}/zookeeper-hosts.yml
    EOA
  }

  # Run Ansible
  provisioner "local-exec" {
    working_dir = var.ansible_workdir
    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
    command = "ansible-playbook -i zookeeper-hosts.yml zookeeper.yml -u ubuntu"
  }
}

resource "yandex_compute_instance" "kafka" {
  count = 3
  name = "kafka-${count.index+1}"
  zone = "ru-central1-a"
  hostname = "kafka-${count.index+1}"
  platform_id = "standard-v2"

  resources {
    cores  = 2
    memory = 4
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size = 15
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.internal-a.id
    nat       = true
  }

  metadata = {
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
  }

  # remote-exec will wait for ssh up and running, after that local-exec will come into play
  # XXX By default requires SSH agent to be running
  provisioner "remote-exec" {
    inline = ["# Connected!"]
    connection {
      host = self.network_interface.0.nat_ip_address
      user = "ubuntu"
    }
  }
}

resource "null_resource" "kafka_deploy" {
  depends_on = [null_resource.zookeeper_deploy]

  triggers = {
    zookeeper_nodes_ids = "${join(",", yandex_compute_instance.zookeeper[*].id)}"
    kafka_nodes_ids = "${join(",", yandex_compute_instance.kafka[*].id)}"
  }

  # Generate Ansible inventory file
  provisioner "local-exec" {
    command = <<-EOA
    echo "${templatefile("${path.module}/templates/ansible_inventory_kafka.yml.tpl", { zookeeper_nodes = yandex_compute_instance.zookeeper[*], kafka_nodes = yandex_compute_instance.kafka[*]})}" > ${var.ansible_workdir}/kafka-hosts.yml
    EOA
  }

  # Run Ansible
  provisioner "local-exec" {
    working_dir = var.ansible_workdir
    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
    command = "ansible-playbook -i kafka-hosts.yml kafka.yml -u ubuntu"
  }
}
