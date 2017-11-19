# Example for using Kafka inside Cloud Foundry

This application is a demonstration of using the https://github.com/edenhill/librdkafka C library as a dependency for a Kafka client library (https://github.com/confluentinc/confluent-kafka-go/kafka in this example).

We use the apt-buildpack to configure and install `librdkafka-dev` package from Confluent's Apt repository. See the `apt.yml` file:

```yaml
---
keys:
- http://packages.confluent.io/deb/3.3/archive.key
repos:
- deb [arch=amd64] http://packages.confluent.io/deb/3.3 stable main
packages:
- librdkafka-dev
```

To use the `apt` buildpack AND the `go` buildpack together, we use the `multi` buildpack. See the `manifest.yml`:

```yaml
applications:
- name: librdkafka-demo-app-cloudfoundry
  buildpack: https://github.com/cloudfoundry/multi-buildpack
  ...
```

And `multi-buildpack.yml` file:

```yaml
buildpacks:
- https://github.com/cloudfoundry/apt-buildpack
- https://github.com/cloudfoundry/go-buildpack
```

To deploy this example app against Stark & Wayne Kafka, run:

```
cf create-service starkandwayne-kafka topic status-topic
cf push librdkafka-demo-app-cloudfoundry --no-start
cf bind-service librdkafka-demo-app-cloudfoundry status-topic
cf restart librdkafka-demo-app-cloudfoundry
cf logs librdkafka-demo-app-cloudfoundry --recent
cf logs librdkafka-demo-app-cloudfoundry
```


During the `cf push` staging process, Cloud Foundry will register the Confluent Apt repository and the install packages. Your output will include the following:

```
  -----> Adding apt keys
  -----> Apt Buildpack version 0.1.1
  gpg: key 41468433: public key "Confluent Packaging <packages@confluent.io>" imported
  gpg: Total number processed: 1
  gpg:               imported: 1  (RSA: 1)
  -----> Updating apt cache
  -----> Adding apt repos
```
