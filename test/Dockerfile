FROM ubuntu:21.10

RUN apt-get update && apt-get install openssh-server sudo -y

RUN useradd --user-group --system --create-home --no-log-init test

RUN echo 'test:test' | chpasswd

USER test

RUN mkdir /home/test/.ssh && touch /home/test/.ssh

COPY ./authorized_keys /home/test/.ssh/authorized_keys

USER root

RUN service ssh start

EXPOSE 22

WORKDIR /home/test

CMD [ "/usr/sbin/sshd", "-D" ]
