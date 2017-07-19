import proto.match_pb2 as proto
from google.protobuf.json_format import MessageToJson
import plyvel, web, json, struct

# Convert bytes to numerical match id
def to_id(raw):
    return int(raw.encode('hex'), 16)

# Convert string match id into byte array (leveldb key)
def to_key(string_id):
    return struct.pack('>q', int(string_id))

urls = (
    '/account/([0-9]+)', 'by_acct',
    '/match/([0-9]+)', 'by_match',
    '/accounts', 'list_acct',
)

app = web.application(urls, globals())

## List all known accounts by summoner name. Returns a mapping between account
## ID and summoner name. Note that this is likely going to return a ton of data,
## so be ready. Your browser might not like it if you're opening it there.
##
## Requires a full iteration of the database.
class list_acct(object):
    def GET(self):
        db = plyvel.DB('matches/db')
        accounts = {}

        for mid, raw in db:
            match = proto.Match()
            match.ParseFromString(raw)

            for participant in match.Participants:
                accounts[participant.AccountID] = participant.SummonerName

        db.close()
        return json.dumps(accounts)

## Get all known games for a particular account. Can be used in conjunction with
## the list_acct endpoint to get match data for a particular summoner name.
##
## Requires a full iteration of the database.
class by_acct(object):
    def GET(self, raw_id):
        target_id = long(raw_id) # parse from string

        db = plyvel.DB('matches/db')

        acct = []

        for mid, raw in db:
            match = proto.Match()
            match.ParseFromString(raw)

            for participant in match.Participants:
                if participant.AccountID == target_id:
                    acct.append(json.loads(MessageToJson(match)))

        db.close()
        return json.dumps(acct)

## Get details about a specific match. Very fast lookups.
class by_match(object):
    def GET(self, raw_id):
        db = plyvel.DB('matches/db')

        raw_match = db.get(to_key(raw_id))
        match = proto.Match()
        match.ParseFromString(raw_match)

        db.close()

        return MessageToJson(match)

if __name__ == '__main__':
    app.run()
