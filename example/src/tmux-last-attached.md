---
title: Tmux session_last_attached
date: 2015-09-23
---

While working on my [session finder](/tmux-session-management), I tought it would be a good idea to list sessions sorted by the last time they were attached.
Tmux has a lot of useful variables for the sessions.
For example, there is the `session_created` variable, which is replaced by the epoch time of when this session was created.
But there is no variable to find out the time that this session was last attached.
I added that and submited a [patch](https://github.com/tmux/tmux/commit/cfabe30becba6f0c54035a29ee61a6a7f3d0cf60) for it.

Now I can list my sessions sorted by the time when they were last attached:


```shell
$ tmux list-sessions -F '#{session_last_attached} #{session_name}' | sort -r
1443039967 blog
1443039956 tmux
1443038611 rails
1443037501 vpn
1443033972 qrencode
1443033969 bender
1443032734 appa
1443013476 today
1443013473 go
1443005180 winterfell
1442995390 chrome-ext
1442941713 linode
1442868891 mp5
1442772955 evc
1442726644 octave
1442567980 fzf
```

I submited a couple of [other](https://github.com/tmux/tmux/commit/16efa8483888e326aed2c05a01b63b45a2b118ef) [patches](https://github.com/tmux/tmux/commit/dc66795e353e1d84c23cb87f4120480a152b43d9) as well.
I use Tmux all the time, I&rsquo;m glad I could contribute something back to it.
