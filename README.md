# What is this?

A small utility that will turn [Plex Media Server](https://plex.tv/) on or off on your OSX system.

# But... why would you want to use that?

I use PMS to stream local videos to a Chromecast device. Chromecast is rather
selective when it comes to playing videos, and prefers H.264. PMS knows about
that, and transcodes on the fly any video that's in a different format. Problem
is, such transcoding eats up a lot of CPU - and the only computer at my home
capable of doing this is my OSX desktop. As a result I run PMS there, instead
of running it on some always-on machine. Given that it's a desktop machine,
I don't necessarily want to have PMS constantly on, staying in the background,
eating up RAM and resources at random intervals. On the other hand, I don't
want to remember to turn it on, every time I want to watch a movie - especially
that the viewing point is in a different room than my desktop.

# The setup

* My always-on device in the network (NAS) hosts the video files.
* The desktop runs `plexup`.
* Upon hitting the right http endpoint, `plexup` mounts the volume with video
  files and starts up PMS.

# Ye gods, you're lazy.

Yup. Laziness is a virtue in my trade :)

# How to install

Get the source, adjust the parameters (port, mount command, etc.). Compile
`plexup.go`, put the binary in some safe place. Hit `/on` endpoint, watch
things click to life.

# The nasty bits and rough edges

* It's a hack I've written because I'm lazy. No guarantees that it'll work
  properly for you. If it doesn't work, tinker. You can contact me about this,
  but chances are you're hitting some setup-specific snag, and I won't be able to
  help you. C'est la vie.
* `plexup` registers itself with Bonjour Sleep Proxy on startup. This means you
  should be able to hit the HTTP endpoints even if machine is sleeping - it'll
  wake up on network traffic. `Wake for network access` has to be turned on in
  `System Preferences - Energy Saver`.
* I had rather poor experience with Plex Media Server (0.9.9.7.429-f80a8d6)
  struggling with machine trying to go to sleep *during* transcoding. You can
  read about the details on [Plex forums](https://forums.plex.tv/index.php/topic/107735-narcolepsy-problems/).
  As a result, you can see `caffeinate` and `pmset` being used during the
  startup sequence for PMS. I've arrived at the use of those utilities and their
  parameters in purely experimental manner. Things might work completely
  different for you. You have been warned.
* This is my first piece of Go ever. It most likely is rather clumsy :)
