FROM ubuntu:21.10

RUN apt-get update && apt-get install openssh-server sudo -y

RUN mkdir -p /home/test/.ssh && \
  touch /home/test/.ssh/authorized_keys && \
  useradd -d /home/test test && \
  usermod -aG sudo test && \
  chown -R test:test /home/test/ && \
  chown test:test /home/test && \
  chmod 700 /home/test/.ssh && \
  chmod 644 /home/test/.ssh/authorized_keys && \
  sudo chsh -s /bin/bash test

RUN echo 'test:test' | chpasswd

RUN service ssh start

EXPOSE 22

WORKDIR /home/test

CMD [ "/usr/sbin/sshd", "-D" ]
