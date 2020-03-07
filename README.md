# GoHome
Have you guys seen those cool looking single filed homepages on Linux communites? It looks really cool, but it's just a shallow shell
which only can let you go to pre-setted internet pages. GoHome in other hand, can do much more, such as opening applications,
shutting your system down, or fetching your system data to your local homepage. Not only that, GoHome also comes with 
easy setup system that does not require any tidious html coding processes, just simple and nice entrie lists-you can customize css as you desire. 

![](https://github.com/SeungheonOh/GoHome/blob/master/img/gohome.jpg)
Theme from [Tanish](https://gitlab.com/Tanish2002/dot-files), Thanks!

# Installation
GoHome is only written with go standard libraries! So you can just download ordinary go compiler, run it!
You can also add gohome to system service so it starts up automatically on your system's startup.

# Example Entrie File
```
!Application
terminal Launch('st')
chrome Launch('google-chrome-stable')
htop Launch('st -e htop')
vim Launch('st -e vim')
code Launch('code')
pipes Launch('st -e pipes.sh')
cowsayToConsole console.log(Launch('cowsay hello world'))

!Website
reddit Launch('xdg-open https://www.reddit.com')
google Launch('xdg-open https://www.google.com')
ArchWiki Launch('xdg-open https://wiki.archlinux.org')
*nix Launch('xdg-open https://www.reddit.com/r/unixporn')

!FunStuff
youtube Launch('xdg-open https://www.youtube.com')
DankMemes Launch('xdg-open https://www.reddit.com/r/dankmemes')
MoonBuggy Launch('st -e moon-buggy')

+CPU: ${Launch('bash -c ./sample_scripts/cpu.sh')}%
+Memory: ${Launch('bash -c ./sample_scripts/memory.sh')}%

```

# Options
```
-c Set CSS file
-t Set Template file
-e Set Entrie file
```

#Inspired by
![/u/whyvitamins](https://www.reddit.com/r/unixporn/comments/da3lx5/bspwm_black_and_white/)
and many other *nix porn posts!
