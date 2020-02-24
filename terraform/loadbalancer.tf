data "yandex_lb_target_group" "events_api_tg" {
  name = "events-api-tg"
}


resource "yandex_lb_network_load_balancer" "events_api_lb" {
  name = "events-api-lb"

  listener {
    name = "events-api-listener"
    port = 80
    target_port = 8080
    external_address_spec {
      ip_version = "ipv4"
    }
  }

  attached_target_group {
    target_group_id = data.yandex_lb_target_group.events_api_tg.id

    healthcheck {
      name = "http"
      http_options {
        port = 8080
        path = "/status"
      }
    }
  }
}

output "ls" {
  value = [
    for v in yandex_lb_network_load_balancer.events_api_lb.listener:
    v.external_address_spec.0.address if v.name == "events-api-listener"
  ][0]
}
