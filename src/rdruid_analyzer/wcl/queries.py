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
                actors(type: "Player") {
                    id
                    name
                    type
                    subType
                    server
                }
            }
        }
    }
}
"""

EVENTS_QUERY = """
query($code: String!, $startTime: Float!, $endTime: Float!, $sourceID: Int!) {
    reportData {
        report(code: $code) {
            events(
                startTime: $startTime,
                endTime: $endTime,
                sourceID: $sourceID,
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
