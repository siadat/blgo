---
title: My experience with Gogs
date: 2015-09-07
---

Until yesterday, we used GitLab at work. GitLab is a great project and it has tones of features. But our GitLab server is a small virtual machine with only 2GB of RAM. We had constant issues with high RAM usage, so much that it would reject some pushes and fetches.

Last night we installed Gogs as a replacement for GitLab. [Gogs](http://gogs.io) is a GitHub-like self-hosted git service written in Go.

The installation was very painless.  We decided to use the existing domain, url and ports, so the team members wouldn&rsquo;t have to update their remote urls.  The most time consuming step was the migration of existing GitLab repositories. I placed the bare repositories created by GitLab and used the migration tool built in Gogs.

-----

After migrating existing repositories, I would receive a 500 error when viewing some of the repositories.
There was an error in the `GetBranch` function.
Gogs did not expect migrated repositories to be empty.
There are many ways to find out if a repository is empty. Using either the `log` command:


```shell
$ git log -1
fatal: bad default revision 'HEAD'
$ echo $?
128
```


or the `rev-parse` command:


```shell
$ git rev-parse --verify HEAD
fatal: Needed a single revision
$ echo $?
128
```


I fixed the error with the help of @Unknown and the [patch](https://github.com/gogits/gogs/pull/1589) is now merged on the develop branch of Gogs.

-----

The overall experience was very positive. The project maintainer is very welcoming and supportive. I highly recommend using Gogs and making contributions to it.
