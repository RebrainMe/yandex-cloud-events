data "yandex_compute_image" "container-optimized-image" {
  family    = "container-optimized-image"
}

resource "yandex_compute_instance_group" "events_api_ig" {
  name               = "events-api-ig"
  service_account_id = yandex_iam_service_account.instances.id

  instance_template {
    platform_id = "standard-v2"
    resources {
      memory = 2
      cores  = 2
    }
    boot_disk {
      mode = "READ_WRITE"
      initialize_params {
        image_id = data.yandex_compute_image.container-optimized-image.id
        size = 10
      }
    }
    network_interface {
      network_id = yandex_vpc_network.internal.id
      subnet_ids = [yandex_vpc_subnet.internal-a.id, yandex_vpc_subnet.internal-b.id, yandex_vpc_subnet.internal-c.id]
      nat = true
    }

    metadata = {
      docker-container-declaration = file("spec.yml")
      ssh-keys = "ubuntu:${file("~/.ssh/id_rsa.pub")}"
    }
    service_account_id = yandex_iam_service_account.docker.id
  }

  scale_policy {
    fixed_scale {
      size = 3
    }
  }

  allocation_policy {
    zones = ["ru-central1-a", "ru-central1-b", "ru-central1-c"]
  }

  deploy_policy {
    max_unavailable = 1
    max_creating    = 1
    max_expansion   = 1
    max_deleting    = 1
  }

  load_balancer {
    target_group_name = "events-api-tg"
  }
}
