
# Outproxys.i2p

**Outproxys.i2p** is a project for maintaining a public database of outproxies in the I2P network.  

An **outproxy** is a proxy inside I2P that allows users to access the clearnet from the I2P network.  

## Features
- Add and list outproxies  
- Track uptime of outproxies  
- Easy integration with I2Pd  

## Build instructions

1. Install **Go ≥ 1.25.1**  
2. In the project root, run:

```bash
   go mod tidy
   go build
```

3. Edit the `config.yaml` file to match your setup
4. Run the server:

   ```bash
   ./outproxys
   ```

## Docker usage

For Docker usage, see [README\_docker.md](README_docker.md).


