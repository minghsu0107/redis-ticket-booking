request = function()
    wrk.headers["X-User-Id"] = os.time()*10000+math.random(1, 10000)
    path = "/ticket"
    return wrk.format("GET", path)
end