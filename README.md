<img src="images/logo.png" width="300">

[![ohmyyaml](./images/operator.png)](https://youtu.be/08O9eLJGQRM "Operators")

# cnskunkworks-operator

This repository provides a "beyond-the-basics" approach to implementing an operator without Operator SDK/Kubebuilder.
The ambition it will help you understand the mechanics of what Operators do and don't actually do.

TL;DR Operators watch CRUD for specific resources and interact with the Kubernetes API.

Learn more about Operators [here](https://github.com/cncf/tag-app-delivery/blob/master/operator-wg/whitepaper/Operator-WhitePaper_v1-0.md#)

## Pre-requisites

- This requires a connection to a kubernetes API ( For testing this could be Kind/Minikube etc)

