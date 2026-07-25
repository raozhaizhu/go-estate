wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["X-Device-ID"] = "test"
wrk.headers["User-Agent"] = "test"
wrk.body = '{"username": "Bob", "password": "12345678"}'

-- local has_printed = false
-- response = function(status, headers, body)
--         if not has_printed then
--                 print("HTTP Status: " .. status)
--                 print("Body: " .. body)
--                 has_printed = true
--         end
-- end