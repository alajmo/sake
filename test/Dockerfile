FROM ubuntu:latest

RUN apt-get update && apt-get install openssh-server sudo -y

RUN useradd --user-group --system --create-home -s /bin/bash --no-log-init test

RUN echo 'test:test' | chpasswd

USER test

RUN mkdir /home/test/.ssh && touch /home/test/.ssh

COPY --chown=test:test ./authorized_keys /home/test/.ssh/authorized_keys

USER root

RUN usermod -aG sudo test
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

RUN service ssh start

WORKDIR /home/test

CMD [ "/usr/sbin/sshd", "-D" ]
