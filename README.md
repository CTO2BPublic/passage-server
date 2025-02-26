# Overview

## What is Passage?
Passage Server is a powerful, open-source access control management solution built in Go. It provides a centralized portal for managing and automating role-based access across multiple platforms and cloud services. Designed with flexibility and scalability in mind.

![Passage UI](assets/passage-ui.webp)

Head to [Passage documentation](https://cto2bpublic.github.io/passage/) to know more.

## Why Passage?
- **Open Source**: Free and community-driven.
- **Go-Powered**: Efficient and performant.
- **Modular**: Easily extendable with new providers.

## Getting started
### Quick start

## How it works
1. **Define Roles**: Create roles that map to specific groups on various platforms.
2. **Request Access**: Users can request access through the Passage Server portal.
3. **Grant & Revoke**: Access can be granted for a limited time and automatically revoked upon expiration.
4. **Multiple Providers**: Use different identity providers (e.g., AWS IAM, GitLab, Google Workspace) through a standardized provider interface.

## Use Cases

- **Engineering Teams**: Manage temporary access for developers across multiple cloud platforms.
  
- **Security Compliance**: Enforce least-privilege access with automatic revocation.
  
- **Multi-Cloud Management**: Simplify role management across diverse platforms.



## Architecture
## Features

- **Unified Access Management**: Manage roles and permissions across multiple platforms like AWS, Google Workspace, and GitLab from a single portal.
  
- **Provider Interface**: Leverage a modular provider system to extend support for various identity platforms easily.
  
- **Role Mapping**: Define roles (e.g., `pu-user`) that map to multiple groups across different platforms.
  
- **Temporary Access**: Grant time-limited access with automatic expiration to reduce over-permissioning.
  
- **User-Friendly Portal**: Web-based portal displaying available roles and their corresponding access mappings.
  
- **Scalable & Secure**: Built in Go for high performance and designed with secure best practices.
