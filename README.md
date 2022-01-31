# Self-Search API

This is a go-lang server allowing a user to manage a set of urls and then search over the content those urls link to.

### Running Locally

1. git clone https://github.com/caltonji/SelfSearch.git
2. ls SelfSearch
3. Create IAM role with access to Firestore.
4. export GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/service-account-file.json"
5. go run .

### Deploying

1. docker buildx build --platform linux/amd64 --push --tag us-west1-docker.pkg.dev/precise-cabinet-280004/personal-images/self-search -f Dockerfile .
2. Deploy to Cloud Run Manually

Deployed publicly at https://kv-multi-g6wckimtwq-uc.a.run.app (Last deployed 1/30/22. No active maintenance.)

# API Documentation

All requests are associated with a `user_id`. When managing urls or searching, this parameter is passed to identify which list of urls is being managed.  There is no validation on the server, we trust that the user issuing a request is doing it with their own `user_id`.

## Get Urls

**URL** : `/urls`

**Method** : `GET`

**URL Params**

**Required** : `user_id`

**Response Body**
The returned content can be used for display.  The `id` field is needed for the DELETE operation.

## Example

`GET /urls?user_id=chris`
Ex: https://kv-multi-g6wckimtwq-uc.a.run.app/urls?user_id=chris
```json
{
    "urls": [
        {
            "url": "https://www.theatlantic.com/magazine/archive/2021/07/america-drinking-alone-problem/619017/",
            "created_at": "2022-01-31T06:24:35.648318Z",
            "title": "America Has a Drinking Problem",
            "image": "cdn.theatlantic.com/thumbor/NMhXVs0MqftLlyhAtxSfkpVUUbk=/0x0:1800x938/960x500/media/img/2021/05/BOB_Julian_Drinking_HPcrop-1/original.jpg",
            "id": "ed5085f1-81ea-434b-a48f-c359aa417642"
        },
        {
            "url": "https://foodwishes.blogspot.com/2013/10/bolognese-sauce-hip-hip-hazan.html",
            "created_at": "2022-01-31T05:59:37.34214Z",
            "title": "Bolognese Sauce – Hip Hip Hazan!",
            "image": "3.bp.blogspot.com/-6p8Asw3ont0/UmauvVGHvbI/AAAAAAAAPV0/H-PecJmSfzc/w1200-h630-p-k-no-nu/IMG_1893.JPG",
            "id": "352b4057-a6cf-4e6f-b22d-6adb94f60331"
        }
    ]
}
```

## Search

**URL** : `/search`

**Method** : `GET`

**URL Params**

**Required** : `user_id`
**Required** : `q`

**Response Body**
`occurrences` is the number of time `q` appears in either the title or the content of the url. `example_text` is a snippet of the content in which `q` appears

`GET /search?user_id=chris&q=bolognese`
Ex: https://kv-multi-g6wckimtwq-uc.a.run.app/search?user_id=chris&q=bolognese
```json
[
    {
        "url": "https://www.bonappetit.com/recipe/bas-best-bolognese",
        "title": "BA's Best Bolognese",
        "image": "assets.bonappetit.com/photos/5c2f8fe26558e92c8a622671/16:9/w_1280,c_limit/bolognese-1.jpg",
        "occurrences": 9,
        "example_text": "es beautifully for M-F. But the best thing about this simple recipe is that it tastes just like the Bolognese my Italian friend makes, because he \"could not find the food mia madre makes in America\". He learne"
    },
    {
        "url": "https://foodwishes.blogspot.com/2013/10/bolognese-sauce-hip-hip-hazan.html",
        "title": "Bolognese Sauce – Hip Hip Hazan!",
        "image": "3.bp.blogspot.com/-6p8Asw3ont0/UmauvVGHvbI/AAAAAAAAPV0/H-PecJmSfzc/w1200-h630-p-k-no-nu/IMG_1893.JPG",
        "occurrences": 4,
        "example_text": "\nThis bolognese sauce is dedicated to the late, great\nMarcella Hazan, who passed away in September, at the age of 8"
    }
]
```

## Add Url

**URL** : `/url`

**Method** : `POST`

**Request Body Type** : `JSON`
**Request Body Fields**

**Required** : `url`
**Required** : `user_id`


`POST /url`
Request Body:
```json
{
    "url": "https://www.theatlantic.com/magazine/archive/2021/07/america-drinking-alone-problem/619017/",
    "user_id": "chris"
}
```

## Delete Url

**URL** : `/url`

**Method** : `DELETE`

**URL Params**

**Required** : `user_id`

`DELETE /url?id=87481c71-aeb1-49ba-9062-6d87eb5d94a0`
