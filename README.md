SSH Agent Downloader
====================

This is an example SSH Agent implementation which dumps any remotely injected keys out to disk.

It was written to work around a security solution where by SSH keys are generated and injected
on a host that is connected to via SSH with key forwarding enabled.

Sadly a number of applications, such as IntelliJ's deployment functionality doesn't work
via the native SSH binary (such as their git support does) and lacks SSH agent support within
their implementation.

Important Notes
---------------

#### The dumped SSH key is *not* encrypted.

Support could be added to provide key encryption, but this would require handling the secret
within the SSH agent implementation.

As this was written initially to be transparent for the user with some assumptions of the
environment (encrypted disk, secure permissions, password requirements, temporary keys etc) it
was decided that this is overall no significantly less secure than having an unlocked SSH key available within the environment, as is the case in a 'normal' SSH setup.

With that said, the key is unencrypted, so if you loose it there is no protection;
at least with an encrypted key there is that single layer of protection, providing a short
time window to revoke the key.

#### stdout/stderr/stdin is proxied

All communications from the user are proxied via this process.
There is a potential that the executable could intercept or change commands,
this is not an explicit intention but is a possibility.

The binary could be run stand-alone, then an un-wrapped SSH call made pointing to the SSH
Agent socket file; to make this easy for the end user a wrapped call was chosen.

#### Key support

We only support RSA key types.

This would be easy to extend for other types, the single type was chosen due to the
requirements at implementation time.

#### Location

The unencrypted key will be written out to `~/.ssh/id_rsa_temporary` as well as being added
to your normal ssh agent via `ssh-add`.

This information is outputted from the process at run-time.

#### Example

```$ ./ssh_agent_download
ssh_agent_download: connecting to localhost

Password:
Two Factor Token:
ssh_agent_download: saved key to /Users/unicorn/.ssh/id_rsa_temporary

Connection to localhost closed.

ssh_agent_download: adding /Users/unicorn/.ssh/id_rsa_temporary to keychain
```


#### Development

* `go fmt`
* `go build`
* `./ssh_agent_download`
