# Entitlement Format

Entitlements are a way to give users different permissions to the bot.

## Bitfield

Entitlements are stored as a bitfield in the database. The bitfield is a 16-bit integer, where each bit represents a different entitlement.

| Bit  | Name      | Description                               |
|------|-----------|-------------------------------------------|
| -1   | None      | No entitlements                           |
| 0    | Admin     | User has admin permissions                |
| 1    | Moderator | User has moderator permissions            |
| 2    | Developer | User has developer permissions            |
| 3    | Premium1  | User has premium 1 entitlement            |
| 4    | Premium2  | User has premium 2 entitlement            |
| 5    | Premium3  | User has premium 3 entitlement            |
| 6    | WFM Staff | This role is for staff of Warframe Market |
| 7-15 | Reserved  | Reserved for future use                   |

## Entitlements

| Entitlement | Items per search | Price Alerts | Pre-Entitlement |
|-------------|------------------|--------------|-----------------|
| None        | 3                | 2            | 0               |
| Admin       | 6                | 5            | 0               |
| Moderator   | 6                | 5            | 0               |
| Developer   | 3                | 5            | 0               |
| Premium1    | 6                | 5            | 5 seconds       |
| Premium2    | 8                | 8            | 10 seconds      |
| Premium3    | 10               | 10           | 30 seconds      |
| WFM Staff   | 10               | 10           | 60 seconds      |
