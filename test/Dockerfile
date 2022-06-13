FROM ubuntu:21.10

RUN apt-get update && apt-get install openssh-server sudo -y

RUN useradd -rm -d /home/ubuntu -s /bin/bash -g root -G sudo -u 1000 test

RUN  echo 'test:test' | chpasswd

RUN service ssh start

EXPOSE 22

CMD ["/usr/sbin/sshd","-D"]
