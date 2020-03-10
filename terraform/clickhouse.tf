resource "yandex_mdb_clickhouse_cluster" "events_ch" {
  name        = "sharded"
  environment = "PRODUCTION"
  network_id  = yandex_vpc_network.internal.id

  access {
    data_lens = true
  }

  clickhouse {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 15
    }
  }

  zookeeper {
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  database {
    name = "events"
  }

  user {
    name     = "events"
    password = "password"
    permission {
      database_name = "events"
    }
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-a"
    subnet_id  = yandex_vpc_subnet.internal-a.id
    shard_name = "shard1"
    assign_public_ip = true
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = yandex_vpc_subnet.internal-b.id
    shard_name = "shard1"
    assign_public_ip = true
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-b"
    subnet_id  = yandex_vpc_subnet.internal-b.id
    shard_name = "shard2"
    assign_public_ip = true
  }

  host {
    type       = "CLICKHOUSE"
    zone       = "ru-central1-c"
    subnet_id  = yandex_vpc_subnet.internal-c.id
    shard_name = "shard2"
    assign_public_ip = true
  }
}
