# feedsync

This CLI application provides the ability to push new RSS/Atom feed entries to your Pocket.

## Usage

```
NAME:
   feedsync - synchronize RSS/Atom feeds to your Pocket

USAGE:
   feedsync [global options] command [command options] [arguments...]

COMMANDS:
   list          list all subscriptions
   add           begin synchronizing a new feed
   remove        stop synchronizing a feed
   sync          run the sync process
   authenticate  retrieve an access token
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

## Configuring

`feedsync` requires a consumer key to make API calls to Pocket and can be obtained from [here](https://getpocket.com/developer/apps/new). After obtaining a one, an access token is needed. `feedsync` provides the `authentication` command to obtain this. This command must be run on the same computer as the browser used to authorize access. Completing the authorization process will automatically save the consumer key and access token to your local config file. The config file is saved in the `feedsync` directory of your user config directory (e.g. `$HOME/.config/feedsync` on linux).

## Adding, Listing, and Removing Feeds

Run `feedsync add <Link>` to add a feed to the list of tracked feeds. By default, the new feed will not sync until the next time `feedsync sync` is run. To override this behavior, pass the `--sync` flag.

Running `feedsync list` provides an output of the form

```
[1] <Link 1>
[2] <Link 2>
...
[n] <Link n>
```

The values next to the links are used to specify a link in the system.

Run `feedsync remove <Feed Number>` to stop tracking a feed. `<Feed Number>` here is the entry number obtained from the `list` command.

## Syncing

Running `feedsync sync` performs the sync process. If a link has not been synced yet, then only the first entry is added to Pocket. Otherwise, all new entries since the last synced entry will be added.

This application does not provide a method for scheduling syncs. Instead, a job scheduler (such as `cron`) should be used.
