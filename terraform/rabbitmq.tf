resource "yandex_compute_instance" "rabbitmq" {
  count = 3
  name = "rabbitmq-${count.index+1}"
  zone = "ru-central1-a"
  hostname = "rabbitmq-${count.index+1}"
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

resource "null_resource" "rabbitmq_deploy" {
  triggers = {
    rabbitmq_nodes_ids = "${join(",", yandex_compute_instance.rabbitmq[*].id)}"
  }

  # Generate Ansible inventory file
  provisioner "local-exec" {
    command = <<-EOA
    echo "${templatefile("${path.module}/templates/ansible_inventory_rabbitmq.yml.tpl", { rabbitmq_nodes = yandex_compute_instance.rabbitmq[*]})}" > ${var.ansible_workdir}/rabbitmq-hosts.yml
    EOA
  }

  # Run Ansible
  provisioner "local-exec" {
    working_dir = var.ansible_workdir
    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
    command = "ansible-playbook -i rabbitmq-hosts.yml rabbitmq.yml -u ubuntu"
  }
}
