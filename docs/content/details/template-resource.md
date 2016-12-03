---
date: 2016-12-03T14:20:57+01:00
next: /details/exec-mode
prev: /details/
title: template resource
toc: true
weight: 5
---

A template resource in remco consists of the following parts:

  - **one optional exec command.**
  - **one or many templates.**
  - **one or many backends.** 
  
{{% notice note %}}
Please note that it is not possible to use the same backend more than once per template resource.
It is for example not possible to use two different redis servers.
{{% /notice %}}