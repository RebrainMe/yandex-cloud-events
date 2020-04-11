# Variables
variable "ansible_workdir" {
  type = string
  description = "Path to Ansible workdir where provisioner tasks are located (i.e. ../ansible)"
}

## Yandex Cloud
variable "yc_token" {
  type = string
  description = "Yandex Cloud API key"
}
variable "yc_region" {
  type = string
  description = "Yandex Cloud Region (i.e. ru-central1-a)"
}
variable "yc_cloud_id" {
  type = string
  description = "Yandex Cloud id"
}
variable "yc_folder_id" {
  type = string
  description = "Yandex Cloud folder id"
}

variable "cf_email" {
  type = string
  description = "Cloudflare email"
}

variable "cf_token" {
  type = string
  description = "Cloudflare api token"
}

variable "cf_zone_id" {
  type = string
  description = "CF zone id"
}

#-----

# Provider
provider "yandex" {
  token = var.yc_token
  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
  zone      = var.yc_region
}

provider "cloudflare" {
  email   = var.cf_email
  api_key = var.cf_token
}
