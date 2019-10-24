# Go service discovery example

This is an example how you can use service discovery with consul and traefik for a go microservice.

## Usage

Clone the repository and use the makefile:

```sh
# Will start infrastucture and services
make start

# Will start infrastucture
make start-infra

# Will start services
make start-services


# Will stop infrastucture and services
make stop

# Will stop infrastucture
make stop-infra

# Will stop services
make stop-services
```
After you have started the containers, you can use the following urls to get access to the services

- Consul: http://localhost:8500/ui/dc1/services
- Traefik: http://localhost:8080/dashboard/
- Greetring-Service: http://localhost:8098/api/greeting/v1/hello/
- User-Service: http://localhost:8099/api/user/v1/hello/

The greeting service also makes an api call to the user service, just as an example.

### Go template

In the templates folder is the consul configuration for each service. To reduce duplicated code, I use `go generate` to copy this file to each service.

-------

If you have any improvements, feel free to open an issue or create a pull request.
