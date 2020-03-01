# Managed by Terraform, don't change by hand
all:
  vars:
    sansible_zookeeper_hosts:
%{for node in zookeeper_nodes}
      - ${node.network_interface.0.ip_address}
%{endfor}
  children:
    zookeeper:
      hosts:
%{for node in zookeeper_nodes}
        ${node.name}:
          ansible_host: ${node.network_interface.0.nat_ip_address}
          private_addr: ${node.network_interface.0.ip_address}
          sansible_zookeeper_id: '${index(zookeeper_nodes, node) + 1}'
%{endfor}
