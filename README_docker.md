## How to use Docker for this project

- open config.yaml
- change host on 0.0.0.0
- change an donation address and interval
- Run `docker-compose build`
- Run `docker-compose up`
- Get your key `docker exec -it i2pd cat /home/i2pd/data/outproxy.dat > key.dat`
- Get your addres `docker run --rm -it -v ${PWD}:/workdir justinhimself/i2pd-tools keyinfo key.dat`
