resource "cloudflare_record" "eye" {
  zone_id = var.cf_zone_id
  name    = "eye"
  value   = yandex_compute_instance.monitoring.network_interface.0.nat_ip_address
  type    = "A"
  ttl     = 1
  proxied = true
}

resource "cloudflare_record" "build" {
  zone_id = var.cf_zone_id
  name    = "build"
  value   = yandex_compute_instance.build.network_interface.0.nat_ip_address
  type    = "A"
  ttl     = 300
  proxied = false
}

resource "cloudflare_record" "events" {
  zone_id = var.cf_zone_id
  name    = "events"
  value   = [
              for v in yandex_lb_network_load_balancer.events_api_lb.listener:
              v.external_address_spec.0.address if v.name == "events-api-listener"
            ][0]
  type    = "A"
  ttl     = 1
  proxied = true
}

