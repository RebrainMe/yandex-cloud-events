data "yandex_compute_image" "ubuntu_image" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "build" {
  name = "build"
  zone = "ru-central1-a"
  hostname = "build"
  platform_id = "standard-v1"

  resources {
    cores  = 1
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size = 10
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

  provisioner "local-exec" {
    working_dir = var.ansible_workdir
    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
    command = "ansible-playbook -u ubuntu -i '${self.network_interface.0.nat_ip_address},' docker.yml"
  }
}

resource "yandex_compute_instance" "monitoring" {
  name = "monitoring"
  zone = "ru-central1-a"
  hostname = "monitoring"
  platform_id = "standard-v1"

  resources {
    cores  = 1
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size = 10
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

  provisioner "local-exec" {
    working_dir = var.ansible_workdir
    environment = {
      ANSIBLE_HOST_KEY_CHECKING = "False"
    }
    command = "ansible-playbook -u ubuntu -i '${self.network_interface.0.nat_ip_address},' monitoring.yml"
  }
}
