resource "yandex_kubernetes_cluster" "kuber-cluster" {
  name        = "kuber-cluster"

  network_id = yandex_vpc_network.internal.id

  master {
    regional {
      region = "ru-central1"

      location {
        zone      = yandex_vpc_subnet.internal-a.zone
        subnet_id = yandex_vpc_subnet.internal-a.id
      }

      location {
        zone      = yandex_vpc_subnet.internal-b.zone
        subnet_id = yandex_vpc_subnet.internal-b.id
      }

      location {
        zone      = yandex_vpc_subnet.internal-c.zone
        subnet_id = yandex_vpc_subnet.internal-c.id
      }
    }

    version   = "1.14"
    public_ip = true

    maintenance_policy {
      auto_upgrade = false
    }
  }

  service_account_id      = yandex_iam_service_account.instances.id
  node_service_account_id = yandex_iam_service_account.docker.id

  labels = {
  }

  release_channel = "STABLE"
}

resource "yandex_kubernetes_node_group" "node-group-0" {
  cluster_id  = yandex_kubernetes_cluster.kuber-cluster.id
  name        = "node-group-0"
  version     = "1.14"

  instance_template {
    platform_id = "standard-v2"
    nat         = true

    resources {
      memory = 2
      cores  = 2
    }

    boot_disk {
      type = "network-hdd"
      size = 64
    }

    scheduling_policy {
      preemptible = false
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    location {
      zone = "ru-central1-a"
    }

    location {
      zone = "ru-central1-b"
    }

    location {
      zone = "ru-central1-c"
    }
  }

  maintenance_policy {
    auto_upgrade = false
    auto_repair  = true
  }
}

