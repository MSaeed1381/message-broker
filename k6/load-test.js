import {Client, Stream} from 'k6/net/grpc';
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
            preAllocatedVUs: 500, // initial VUs
            maxVUs: 1100, // maximum VUs
        },
    },
    thresholds: {
        'publish_requests': ['count > 100'], // ensure at least 100 publish requests were made
        'fetch_requests': ['count > 100'], // ensure at least 100 fetch requests were made
        'checks': ['rate>0.95'], // ensure at least 95% of checks passed
    },
};

export default () => {
    client.connect('localhost:8000', {
        plaintext: true,
    });

    const publishRequest = {
        subject: 'test-subject',
        body: 'test',
        expirationSeconds: 60,
    };

    const fetchRequest = {
        subject: 'test-subject',
        id: 1, // assuming 1 is a valid message ID for the test
    };

    // Publish
    const publishResponse = client.invoke('broker.Broker/Publish', publishRequest);
    check(publishResponse, {
        'publish status is OK': (r) => r && r.status === grpc.StatusOK,
    });
    publishCounter.add(1);

    // Fetch
    const fetchResponse = client.invoke('broker.Broker/Fetch', fetchRequest);


    console.log(fetchResponse.message.body)


    check(fetchResponse, {
        'fetch status is OK': (r) => r && r.status === grpc.StatusOK,
        'fetch response body is correct': (r) => r.message && r.message.body === 'test',
    });
    fetchCounter.add(1);

    client.close();
    sleep(1); // pause for 1 second between iterations
};
