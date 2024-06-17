# Entitlement Format

Entitlements are a way to give users different permissions to the bot.

## Bitfield

Entitlements are stored as a bitfield in the database. The bitfield is a 16-bit integer, where each bit represents a different entitlement.

| Bit  | Value     | Name      | Description                               |
|------|-----------|-----------|-------------------------------------------|
| -1   | 0         | None      | No entitlements                           |
| 0    | 1         | Admin     | User has admin permissions                |
| 1    | 2         | Moderator | User has moderator permissions            |
| 2    | 4         | Developer | User has developer permissions            |
| 3    | 8         | Premium1  | User has premium 1 entitlement            |
| 4    | 16        | Premium2  | User has premium 2 entitlement            |
| 5    | 32        | Premium3  | User has premium 3 entitlement            |
| 6    | 64        | WFM Staff | This role is for staff of Warframe Market |
| 7-15 | 128-32768 | Reserved  | Reserved for future use                   |

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
