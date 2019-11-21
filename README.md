# TODO

- Example configs? - Start with testing
- Reset git
- Diagram
- Better install instructions
- Begin writing crostini post, reach out to mods to ask permission?

-----------

This is highly experimental.

Code is quite bad at the moment, but interested in feedback mostly on the concept.

The goal is to provide a seamless way to run commands on a remote server while still
having files local for your editor to work with or other fast commands.  This
would improve workflow around running long running commands on remote compute.

As an example
```
rrun npm run test
```

or in ml land

```
rrun python3 train.py
```

# Pre-requisites

## Remote machine

- apt-get install rsync openssh-server nix-shell
- Accepts public key without password from your local machine.

## Local machine

- apt-get install rsync openssh-client
- Add rrun/bin to path
- Install and start daemon (non-root user)
- Directory you run needs rrun.toml and dev.nix

# Current Features

- Rough 2-way sync via rsync (very inefficient, doesn't handle deletes)
- Run command remotely
- Tunnel support (connect a port)


# How it Works

![Architecture](/rrun.svg)


# Planned Features

- Seamless command resume
- Better and more efficient sync setup.
- Kubernetes as compute backend
- Auto configure remote
- Auto shutdown after ttl expiry
- Teardown unused resources
- Auto spin up and tear down of resources


# Example Use Cases

## Machine learning - Training

How to bring trained model back to local?
Robust to disconnect/reconnect

rrun -- python3 train.py

## Intense testing

rrun -- npm run test

## Long project build

rrun -- npm run build

## Collaboration (Expose to outside world)

rrun --demo -- npm run start
