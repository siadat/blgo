---
title: Tmux session management
date: 2015-09-14
---

<style>
  .after-before div:last-child { display:none; }
  .after-before:hover div:first-child { display:none; }
  .after-before:hover div:last-child { display:block; }
  .after-before img { border-radius:5px; width:100%; }
</style>

When using Tmux, I try to follow these rules:

- create a named session for each project and long running process.
- create a window for each task in the session.
- create a pane for closely related actions of each task.
- each window should not include more than 2 panes.

I have written a [session-finder](https://github.com/siadat/session-finder/blob/master/session-finder.bash)
Bash script that utilises [fzf](https://github.com/junegunn/fzf) to quickly
find and create new sessions and switch between them by typing only a few
characters of their name.  A new session is created if no matching session is found.
Example sessions are irc clients, ssh, vpn, and proxy connections.

A problem I encountered when writing session-finder was that fzf returned exit status `0` even when no match was found.
I came up with a simple workaround. I passed `--print-query` to fzf and counted the printed lines.
When there is a match fzf prints:


    <query>
    <selected line>


And when there is no match:


    <query>


It would have been simpler if fzf returned a non-zero exit status when no match is found.

**Update:** I opened an [issue](https://github.com/junegunn/fzf/issues/345) and requested for non-zero exit status, and @junegunn implemented it very quickly. It will be available from 0.10.6.

Back to the session-finder.
Here is an example of how my session-finder looks like:

<div class="after-before">
  <div> <img src="/assets/tmux/1.png"> <br> <i>11 named sessions, 2 windows in the current session.</i> </div>
  <div> <img src="/assets/tmux/2.png"> <br> <i>fuzzy session finder and creator.</i> </div>
</div>

In this screenshot, there are 11 sessions, &ldquo;tmux&rdquo;, &ldquo;appa&rdquo;, &ldquo;crystal&rdquo; etc.
The hover image shows what happens when finder is started is called.
I typed `l` to filter sessions whose name contains this letter.
Below is a list of all available sub-commands.

Print a pretty list of session names to be used by `status-left`:


    bash session-finder.bash status


Start fzf and find or create sessions:


    bash session-finder.bash finder


Switch to the last session:


    bash session-finder.bash last


Switch to the next session:


    bash session-finder.bash next


Switch to the previous session:


    bash session-finder.bash prev

