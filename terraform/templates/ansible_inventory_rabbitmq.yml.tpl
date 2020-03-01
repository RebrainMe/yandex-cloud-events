# Managed by Terraform, don't change by hand
all:
  vars:
    rabbitmq_cluster: true
    rabbitmq_cluster_master: "rabbit@${rabbitmq_nodes[0].hostname}"
    rabbitmq_erlang_cookie_file: /var/lib/rabbitmq/.erlang.cookie
    rabbitmq_plugin_dir: "/usr/lib/rabbitmq/lib/rabbitmq_server-{{ rabbitmq_version }}/plugins"
    rabbitmq_erlang_cookie: "b64XjcaCxudHBoksX2nIC2qPbo3CsOQe"

    rabbitmq_plugins:
      - rabbitmq_management

    rabbitmq_plugins_disabled: []

    rabbitmq_users:
      - user: admin
        password: admin
        tags: administrator

    rabbitmq_users_absent: []

    rabbitmq_version: '3.8.2-1'
    rabbitmq_vhosts: []
    rabbitmq_vhosts_absent: []

  children:
    rabbitmq:
      hosts:
%{for node in rabbitmq_nodes}
        ${node.hostname}:
          ansible_host: ${node.network_interface.0.nat_ip_address}
          private_addr: ${node.network_interface.0.ip_address}
%{endfor}
