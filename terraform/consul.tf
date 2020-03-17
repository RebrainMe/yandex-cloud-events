resource "yandex_compute_instance" "consul" {
  name = "consul"
  zone = "ru-central1-a"
  hostname = "consul"
  platform_id = "standard-v1"

  resources {
    cores  = 1
    memory = 2
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.container-optimized-image.id
      size = 10
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.internal-a.id
    nat       = true
  }

  metadata = {
    docker-container-declaration = file("${path.module}/templates/instance_consul_spec.yml")
    ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
  }

  service_account_id = yandex_iam_service_account.docker.id
}
