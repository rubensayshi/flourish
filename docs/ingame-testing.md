# In-Game Testing Notes

Things to verify in-game to confirm attribution logic.

## SotF + Pandemic Refresh
- Cast Swiftmend → Rejuv (SotF-buffed) → let it tick a few times → refresh Rejuv in pandemic window
- Does the refreshed Rejuv still have the SotF +60% buff, or does it revert to normal?
- This determines if our tag-persistence-through-refresh behavior is correct or a bug
