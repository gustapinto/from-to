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
