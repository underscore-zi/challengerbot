# Challenger Bot

A simple Discord bot that is capable of serving a CTF-style challenges and rewarding roles for successfully solving challenges.

## Commands

 - **challenges** - Will display a list of challenge IDs and their titles
 - **challenge <id>** - Will display the information you need to know to get started on the challenge
 - **hint <id> [number]** -Will display a hint for a challenge, if a challenge has multiple hints they can be accessed by index starting with 1.
 - **commands** - Lists the supported commands and a brief description.

## Configuration

### Bot Creation

To start you will need to login to the Discord Developer Portal and create a new Application with a Bot.

The following permissions are used:

 - Manager Roles - Needed to modify the roles of users who submit a flag
 - Send Messages - Needed to send messages to the chat
 - Embed Links - Needed to link to website relating to a challenge. This can be ignored if you don't configure any challenges with links
 - Attach Files - Needed to attach challenge files. This can be ignored if you do not configure any challenges with attached files
 - Oauth2 Link: discord.com/api/oauth2/authorize?client_id=**YOUR_APPLICATION_CLIENT_ID_HERE**&permissions=268486656&scope=bot
 
Once the bot has been joined to your guild/server note the bot's 

### config.json

Primary configuration for the bot is done using a .json file that is loaded when the bot starts. There is presently no way to update the bot's configuration at runtime. You can refer to ./config.json for an example configuration file. 

All values unless otherwise noted are *strings*
##### Config Keys
**guild** - The Guild ID that bot should run in

**channel** - The Channel ID that the bot should listen in. If blank the bot will be active in all channels. 

**default_color** - This is the hex for the color next to embeds. 
![](http://se.ri0.us/2020-07-23-204257846-4b28a.png)

**command_string** - All commands will be required to start with this string. "!" is a common command string but longer prefixes can be used.

**roles** - This contains a dictionary mapping human readable role names to their respective discord ids. 

The Role ID can be obtained from discord by enabling role highlighting and then typeing `\@RoleNameHere` the backslash is essential as it will cause Discord to print the id rather than working as a highlight. The Role Id is the number between `<@&` and `>`

![](http://se.ri0.us/2020-07-23-200738151-a6081.png)

Example `roles` Value:

```
"roles":{
    "exploit_dev":"655516092680241172",
    "reverse_engineer":"654650040534564884",
    "programmer":"654649713441898498",
    "pentester":"654650193219813376"
},
```

These human readable identifiers are used later in the challenge list to name the roles that a challenge should unlock.

**challenges** - This should contain an array[] of challenge objects.

Each challenge object contains the following fields:

 - **id** - This should be a simple name to identify the challenge to the users and is used by the `challenge <id>` command to find challenges so the identifier should be easy and memorable to type.  
 - **name** - This is the proper title of a challenge and is displayed in challenge embed.
 - **description** - The description is the prompt that should let the user know everything they need to know about the challenge. 
 - **hints** - This is an array[] of strings/hints. 
 - **flag** - Flag submissions are detected by the presence of `{` and `}` in a DM to the bot. The content between the first `{` and the last `}` is the value that is checked against this value in the config. So while all flags must be submitted in the form of FLAG{...} (or any prefix you want instead of FLAG) the flag in the config should not contain the wrapping text.
 - **role** - A single role can be applied per challenge. The value here is the same as the human readable role identifier provided earlier in the `roles` field of the main configuration.

Challenge objects can also contain the follow optional fields

 - **filename** - This field should be the name of a file in the ./files directory that should be served with the challenge. Discord does place limits on file uploads so keep that in mind. 
 - **filetype** - The MIME type of the file.
 - **link** - If this is provided the challenge embed's Title will be clickable and will navigate to this location.
 - **color** - This can be used to override the default color of the embed

#### Example Challenge Configurations

**Challenge containing a `link` field**

```
{
  "id":"sorting",
  "name": "Hacker Sort (Programmer)",
  "description": "Just sort the |_337-ified comma-separated words into alphabetical order in under two seconds.",
  "link": "http://little-canada.org/roguesec/programming.php",
  "flag": "redacted",
  "role": "programmer",
  "hints": [
    "Solve it manually first, then automate that process.",
    "Leetifying words is just replace a letter you know, and replacing it with some other characters that look similar, like m becomes |\\/|, try to figure out what all the replacements are.",
    "Don't forget to send your session id cookie if you are not solving this in the browser."
  ]
}
```
![](http://se.ri0.us/2020-07-23-203739802-76c21.png)

**Challenge containing an attached file (`filename` and `filetype` fields)
```
{
  "id":"fragmented",
  "name": "Fragmented (Reverse Engineer)",
  "description": "There's a binary that would give you the flag, but someone got angry and tore it into seven pieces. Re-assemble it and grab the flag.",
  "flag": "redacted",
  "role": "reverse_engineer",
  "filename": "fragmented.zip",
  "filetype": "application/zip",
  "hints": [
    "Start by trying to figure out what type of file this is.",
    "Most files have what is called a magic number that identifies its file type",
    "Executable files contain information about their layouts maybe you can use that to help"
  ]
}
```
![](http://se.ri0.us/2020-07-23-203938530-ae8e3.png)

**Simple challenge with no optional fields**

```
{
  "id":"bleed",
  "name": "Bleed The Stack (Exploit Developer)",
  "description": "An amateur programmer decides that for his hello world program, he will echo whatever you say. Can you find his mistake?\n\n`nc challenges.0x0539.net 7070`",
  "flag": "redacted",
  "role": "exploit_dev",
  "hints": [
    "Many beginner C programmers make this type of mistake when printing user input to the screen",
    "What function is commonly used in C to print strings to the screen, are any well known to be a source of bugs?"
  ]
}
```


![](http://se.ri0.us/2020-07-23-204059667-4a9b6.png)

### Environment

Two environment variables are necessary:

 - **BOT_TOKEN** - Needs to contain a Discord application bot token. This can be obtained from the Discord developer panel by creating a new application, adding a bot to it, and revealing the bot's token from https://discord.com/developers/applications
 - **BOT_CONFIG_FILE** - This it the absolute path to the config.json file.

 
  






