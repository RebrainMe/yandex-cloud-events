resource "yandex_iam_service_account" "docker" {
  name        = "docker"
  description = "service account to use container registry"
}

resource "yandex_iam_service_account" "instances" {
  name        = "instances"
  description = "service account to manage VMs"
}

resource "yandex_resourcemanager_folder_iam_binding" "editor" {
  folder_id = var.yc_folder_id

  role = "editor"

  members = [
    "serviceAccount:${yandex_iam_service_account.instances.id}",
  ]
}

resource "yandex_resourcemanager_folder_iam_binding" "puller" {
  folder_id = var.yc_folder_id

  role = "container-registry.images.puller"

  members = [
    "serviceAccount:${yandex_iam_service_account.docker.id}",
  ]
}

resource "yandex_resourcemanager_folder_iam_binding" "pusher" {
  folder_id = var.yc_folder_id

  role = "container-registry.images.pusher"

  members = [
    "serviceAccount:${yandex_iam_service_account.docker.id}",
  ]
}
