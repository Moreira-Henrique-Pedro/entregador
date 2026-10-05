FROM node:22-alpine

RUN npm install -g firebase-tools@15.32.1 && firebase setup:emulators:ui

WORKDIR /firebase
COPY firebase.json .

EXPOSE 9099 4000 4400 4500

CMD ["firebase", "emulators:start", "--only", "auth", "--project", "demo-entregador"]
