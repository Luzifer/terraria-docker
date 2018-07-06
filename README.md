[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/terraria-docker)](https://goreportcard.com/report/github.com/Luzifer/terraria-docker)
![](https://badges.fyi/github/license/Luzifer/terraria-docker)
![](https://badges.fyi/github/downloads/Luzifer/terraria-docker)
![](https://badges.fyi/github/latest-release/Luzifer/terraria-docker)

# Luzifer / terraria-docker

This application is intended as a wrapper around the [Terraria](http://terraria.org/) server when running in a Docker container. It ensures `STDIN` is not closed and therefore the Terraria server does not crash when daemonized. Also it works around the issue the server does not save the world after the last player left: For this the log is observed and on a player leaving the server a direct `save` is issued.

Additionally a fifo is created to enable the user to issue commands on the server after starting using `docker exec`.

All those features are used inside my [Terraria server container](https://github.com/luzifer-docker/terraria).
