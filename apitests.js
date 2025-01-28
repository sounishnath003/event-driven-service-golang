import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 60 }, // ramp up to 60 users
        { duration: '1m', target: 100 },  // stay at 100 users for 1 minute
        { duration: '30s', target: 0 },  // ramp down to 0 users
    ],
};

export default function () {
    let payload = JSON.stringify({
        "author": "sounish",
        "title": `Blog title ${Math.random()}`,
        "subtitle": `Subtitle ${Math.random()}`,
        "content": `My blog content ${Math.random()}`
    });

    let params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    let res = http.post('http://localhost:3000/api/create-post-with-kafka', payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time is less than 200ms': (r) => r.timings.duration < 200,
    });

}