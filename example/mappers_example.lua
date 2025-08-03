function map_sales_event(event)
    local op_mapping = {
        ['I'] = 'INSERT',
        ['U'] = 'UPGRADE',
        ['D'] = 'DELETE',
    }

    return {
        ['action'] = op_mapping[event['op']],
        ['index'] = event['table'] .. '/' .. event['row']['id'],
        ['data'] = event['row'],
    }
end

function map_sales_event_with_http(event)
    local http = require("http")
    local json = require("json")

    local op_mapping = {
        ['I'] = 'INSERT',
        ['U'] = 'UPGRADE',
        ['D'] = 'DELETE',
    }

    local res, err = http.get('https://jsonplaceholder.typicode.com/users', {
        ['headers'] = {
            ['Accept'] = 'application/json',
        },
    })

    local users = json.decode(res.body)

    return {
        ['action'] = op_mapping[event['op']],
        ['index'] = event['table'] .. '/' .. event['row']['id'],
        ['data'] = event['row'],
        ['user_id'] = users[1]['id'],
        ['user_name'] = users[1]['name'],
    }
end
