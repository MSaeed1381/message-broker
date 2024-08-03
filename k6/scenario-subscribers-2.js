import { Client, Stream } from 'k6/net/grpc';
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

const publishCounter = new Counter('publish_requests');
const fetchCounter = new Counter('fetch_requests');

const client = new Client();
client.load(['../api/proto'], 'broker.proto');

export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 1000, // number of requests per second
            timeUnit: '1s', // per second
            duration: '20s', // test duration
            preAllocatedVUs: 100, // initial VUs
            maxVUs: 20000, // maximum VUs
        },
    },
    thresholds: {
        'publish_requests': ['count > 100'], // ensure at least 100 publish requests were made
        'fetch_requests': ['count > 100'], // ensure at least 100 fetch requests were made
        'checks': ['rate > 0.95'], // ensure at least 95% of checks passed
    },
};

export default () => {

    client.connect('127.0.0.1:8000', {
        plaintext: true,
        timeout: '120s',
    });

    const publishRequest = {
        subject: 'ali',
        body: 'test',
        expirationSeconds: 60,
    };

    const subscribeRequest = {
        subject: 'ali',
    };

    // Subscribe
    const stream = new Stream(client, 'broker.Broker/Subscribe');

    stream.on('data', (message) => {
        check(message, {
            'message body is correct': (msg) => msg.body === 'test',
        });
        fetchCounter.add(1);
    });

    stream.on('error', (error) => {
        console.error(`Stream error: ${error.message}`);
    });

    stream.write(subscribeRequest)

    sleep(10)

    // Publish
    const publishResponse = client.invoke('broker.Broker/Publish', publishRequest);
    check(publishResponse, {
        'publish status is OK': (r) => r && r.status === grpc.StatusOK,
    });
    publishCounter.add(1);

    sleep(1); // pause for 1 second between iterations

    client.close();
};
