# Managed by Terraform, don't change by hand
all:
  vars:
    sansible_zookeeper_hosts:
%{for node in zookeeper_nodes}
      - ${node.network_interface.0.ip_address}
%{endfor}

    sansible_kafka_zookeeper_hosts:
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

    kafka:
      hosts:
%{for node in kafka_nodes}
        ${node.name}:
          ansible_host: ${node.network_interface.0.nat_ip_address}
          private_addr: ${node.network_interface.0.ip_address}
          sansible_kafka_server_properties:
            num.partitions: 15
            broker.id: ${index(kafka_nodes, node) + 1}
            log.dirs: /home/kafka/data
            listeners: "PLAINTEXT://0.0.0.0:{{ sansible_kafka_port }}"
            advertised.listeners: "PLAINTEXT://${node.network_interface.0.ip_address}:{{ sansible_kafka_port }}"
%{endfor}
