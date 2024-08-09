import requests
import os
import sys

cmd = sys.argv[1]

api_url = "https://api.github.com/repos/DiceDB/dice/issues"
token = os.environ.get("GITHUB_TOKEN")

title = f"Add support for command `{cmd}`"
body = f"""Add support for the `{cmd}` command in DiceDB similar to the [`{cmd}` command in Redis](https://redis.io/docs/latest/commands/command-list/). Please refer to the following commit in Redis to understand the implementation specifics - [source](https://github.com/redis/redis/tree/f60370ce28b946c1146dcea77c9c399d39601aaa).

Write unit and integration tests for the command referring to the tests written in the [Redis codebase 7.2.5](https://github.com/redis/redis/tree/f60370ce28b946c1146dcea77c9c399d39601aaa). For integration tests, you can refer to the `tests` folder. Note: they have used `TCL` for the test suite, and we need to port that to our way of writing integration tests using the relevant helper methods. Please refer to our [tests directory](https://github.com/DiceDB/dice/tree/master/tests).

For the command, benchmark the code and measure the time taken and memory allocs using [`benchmem`](https://blog.logrocket.com/benchmarking-golang-improve-function-performance/) and try to keep them to the bare minimum.
"""

def create_github_issue(title, body):
    headers = {
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github.v3+json"
    }
    
    data = {
        "title": title,
        "body": body
    }
    
    response = requests.post(api_url, json=data, headers=headers)
    if response.status_code == 201:
        print(f"Successfully created issue: {title}")
    else:
        print(f"Failed to create issue: {title}")
        print(f"Status code: {response.status_code}")
        print(f"Response: {response.text}")


if __name__ == "__main__":
    create_github_issue(title, body)
