## Simple Fantasy

A very crude tool that lists the "perfect team" for a Premier League fantasy gameweek. It factors in the following qualities:
1) The estimated "difficulty" of the fixture.
2) A player's form.
3) A player's ICT index.
4) A player's average starts.
5) A player's likelihood of playing.

e.g.

<img src="./img.png" />

### Usage
```
simple-fantasy -gameweek 10
```

#### Player Detail
```
simple-fantasy -gameweek 10 -player Haaland
```

#### Config Option
```
simple-fantasy -gameweek 10 -manager-id {your-manager-id}
```
You can find your manager ID by logging into your FPL account, clicking on "Points" and taking the number in the URL after "entry/".

<img src="./img2.png" />

