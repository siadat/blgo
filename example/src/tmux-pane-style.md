---
title: Colorful Tmux panes
date: 2015-09-16
---

<style>
  img.fullwidth { border-radius:5px; width:100%; }
</style>

Thanks to [this commit](https://github.com/tmux/tmux/commit/ee123c24), we are able to style an individual pane in Tmux.
You will have to build Tmux from source to do that. Try:

```shell
$ tmux select-pane -P 'bg=green'
```

And you will get something like:

<img class="fullwidth" src="/assets/tmux/select-pane.png">

We could use that to highlight an important pane.
I have set key-bindings `M-1` to `M-4` to different background colours, and `M-0` to the default colour:


    bind-key -n 'M-0' select-pane -P 'bg=default'
    bind-key -n 'M-1' select-pane -P 'bg=black'
    bind-key -n 'M-2' select-pane -P 'bg=green'
    bind-key -n 'M-3' select-pane -P 'bg=blue'
    bind-key -n 'M-4' select-pane -P 'bg=red'


Thanks @thomas_adam for mentioning this feature in the #tmux channel.
