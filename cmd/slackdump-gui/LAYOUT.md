# GUI Layout Description

## Kobe User Research Slack Export Tool

### Window Layout

```
┌──────────────────────────────────────────────────────────────┐
│           Kobe User Research Slack Export Tool               │
│                  For Research Purposes Only                   │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Workspace Subdomain:                                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Enter workspace subdomain (e.g., myworkspace)          │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
│  Cookie (d value):                                           │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ ••••••••••••••••••••••••••••••••••••••••••••••••••••   │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
│  Export Year: [2025 ▼]  (Dropdown: 2025, 2024, 2023, etc.)  │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Get Full List of Channels                      │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
├──────────────────────────────────────────────────────────────┤
│  Status: [Authentication and operation status shown here]    │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Channels List (scrollable):                            │ │
│  │                                                         │ │
│  │ • general (C12345)                                      │ │
│  │ • random (C67890)                                       │ │
│  │ • engineering (C11111)                                  │ │
│  │ • marketing (C22222)                                    │ │
│  │ ...                                                     │ │
│  │                                                         │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
├──────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Export Selected Channels                       │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### Features

1. **Branding**: Clear "Kobe User Research Slack Export Tool" title with "For Research Purposes Only" subtitle
2. **Authentication**: 
   - Subdomain field (plain text)
   - Cookie field (password masked)
3. **Date Selection**: Dropdown defaulting to 2025
4. **Get Channels**: Button to authenticate and retrieve all channels
5. **Status Display**: Shows authentication status and operation progress
6. **Channel List**: Scrollable list displaying channel name and ID
7. **Export Button**: Placeholder for future export functionality

### User Flow

1. User enters workspace subdomain (e.g., "kobe-research")
2. User enters their 'd' cookie value (obtained from browser)
3. User selects export year (defaults to 2025)
4. User clicks "Get Full List of Channels"
5. Application authenticates and retrieves channels
6. Channels are displayed in the scrollable list
7. User can click "Export Selected Channels" (currently shows info dialog as placeholder)

### Technical Implementation

- Built with Fyne.io v2.7.1
- Direct integration with slackdump package
- No external binary calls
- Clean, minimal, MVP implementation
- Proper error handling with dialog boxes
- Password masking for cookie field
