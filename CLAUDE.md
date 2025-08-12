## Project Ideas

### RentalManagement Tool
- Goal to build a highly professional and visually appealing rental management application
- Focus on consistent design theme across the entire app (font, color scheme, overall perception)
- Prioritize fast website loading performance

## Sensitive Configuration

### Database Credentials
- External database hosted at tsunami-events.de
- Username: tsweb
- Password: N1KO4cCYnp3Tyf
- Database: TS-Lager

## Project Resources

### Database Templates
- Das aktuelle Datenbank template liegt im root der Codebase und heißt TS-Lager.sql

## Professional Development Mindset

### Software Development Philosophy
- Du bist ein höchstprofessioneller Softwaredeveloper, welcher niemals sagt, dass ein Programm fertig wär oder funktionieren würde, obwohl es das nicht zu 100% tut.

### File Management
- Wenn du temporäre Dateien oder Debug Dateien anlegst, dann löschst du die Dateien sofort, nachdem du sie nicht mehr brauchst. Auch wenn du sowas wie eine neue Version einer vorhandenen Datei anlegst, dann wird die alte sofort gelöscht.

## Development Workflow

### Server Management
- Please restart the Server always, after you made changes, so i dont have to.
- NEVER use the command: pkill server because it closes my tmux session which is NOT WANTED
- ALWAYS build and push the project to docker hub image: nbt4/rentalcore and please make the version tag simple like 1.X and so on.
- and always push your changes to github for possible easy rollbacks. and i mean every time you finished smth. push it to github