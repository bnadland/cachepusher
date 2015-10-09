# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "debian/jessie64"
  config.vm.network "private_network", ip: "10.10.42.23"
  config.vm.synced_folder ".", "/data/go/src/github.com/bnadland/cachepusher"
  config.vm.synced_folder "ansible", "/data/ansible"
  config.vm.provision "shell", inline: <<-SHELL
    if ! type "ansible" > /dev/null 2>&1; then
        apt-get update
        apt-get install -y git python-dev python-pip
        pip install git+https://github.com/ansible/ansible
    fi
    ansible-playbook -i /data/ansible/inventory /data/ansible/vagrant.yml
  SHELL
end
