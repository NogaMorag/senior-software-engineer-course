package outbox

// HeaderMessageID is the record header that carries Message.ID. The relay sets it; every
// consumer reads it and dedupes on it. The Kafka offset is not a dedupe key: a duplicate
// publish gets a new offset.
const HeaderMessageID = "message-id"
