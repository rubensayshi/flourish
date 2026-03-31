FIGHTS_QUERY = """
query($code: String!) {
    reportData {
        report(code: $code) {
            title
            fights {
                id
                name
                kill
                startTime
                endTime
                difficulty
                encounterID
            }
            masterData {
                actors {
                    id
                    name
                    type
                    subType
                    server
                    petOwner
                }
            }
        }
    }
}
"""

EVENTS_QUERY = """
query($code: String!, $startTime: Float!, $endTime: Float!, $sourceID: Int!, $fightIDs: [Int!]) {
    reportData {
        report(code: $code) {
            events(
                startTime: $startTime,
                endTime: $endTime,
                sourceID: $sourceID,
                fightIDs: $fightIDs,
                dataType: All,
                limit: 10000
            ) {
                data
                nextPageTimestamp
            }
        }
    }
}
"""
