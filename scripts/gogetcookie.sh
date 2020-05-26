#!/usr/bin/env bash

eval 'set +o history' 2>/dev/null || setopt HIST_IGNORE_SPACE 2>/dev/null
 touch ~/.gitcookies
 chmod 0600 ~/.gitcookies

 git config --global http.cookiefile ~/.gitcookies

 tr , \\t <<\__END__ >>~/.gitcookies
source.developers.google.com,FALSE,/,TRUE,2147483647,o,git-vraterraformcmbu.gmail.com=1//0fYBpthpHhG4DCgYIARAAGA8SNwF-L9Ir9R57XrRWI9szPkeloSkaiDm4wivtTEFc5f7kVwy_fw4CvP2S-VvTwVCxZZpU1eelEbE
__END__
eval 'set -o history' 2>/dev/null || unsetopt HIST_IGNORE_SPACE 2>/dev/null