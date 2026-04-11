# Template resource

A template resource in remco consists of the following parts:

- **one optional exec command.**
- **one or many templates.**
- **one or many backends.**

!!! note
    It is not possible to use the same backend more than once per template resource. For example, it is not possible to use two different redis servers.
