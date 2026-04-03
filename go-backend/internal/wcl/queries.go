package wcl

const FightsQuery = `
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
`

const EventsQuery = `
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
`

const FightEventsQuery = `
query($code: String!, $startTime: Float!, $endTime: Float!, $fightIDs: [Int!]) {
    reportData {
        report(code: $code) {
            events(
                startTime: $startTime,
                endTime: $endTime,
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
`

const ResourcesQuery = `
query($code: String!, $startTime: Float!, $endTime: Float!, $sourceID: Int!, $fightIDs: [Int!]) {
    reportData {
        report(code: $code) {
            events(
                startTime: $startTime,
                endTime: $endTime,
                sourceID: $sourceID,
                fightIDs: $fightIDs,
                dataType: Resources,
                limit: 10000
            ) {
                data
                nextPageTimestamp
            }
        }
    }
}
`

const DamageTakenTableQuery = `
query($code: String!, $startTime: Float!, $endTime: Float!, $sourceID: Int!, $fightIDs: [Int!], $filterExpression: String) {
    reportData {
        report(code: $code) {
            table(
                startTime: $startTime,
                endTime: $endTime,
                sourceID: $sourceID,
                fightIDs: $fightIDs,
                dataType: DamageTaken,
                filterExpression: $filterExpression
            )
        }
    }
}
`
