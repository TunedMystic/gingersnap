---
title: Diving into SQLite Full-Text Search - The Complete Guide
heading: SQLite Full-Text Search
slug: sqlite-full-text-search
description: In this guide, you'll learn all about SQLite's Full-Text Search feature and how to use it to efficiently search and retrieve data from your database.

category: SQL
image_url: /media/full-text-search.webp
image_alt: SQLite Full-Text Search

pubdate: 2023-01-21
updated: 2023-01-23
featured: true
---

Lorem ipsum dolor sit amet, consectetur adipiscing elit. [Fusce nec tincidunt](https://news.ycombinator.com/) ipsum, posuere hendrerit metus. Proin eu sapien ipsum. Fusce `posuere lectus ut ex` cursus, at maximus felis aliquet. Integer sodales orci non massa pulvinar lobortis.

Maecenas leo ex, ultricies non accumsan nec, [`egestas eget`](https://twitter.com/golang) nulla. Etiam sollicitudin dolor et libero tempor, sed scelerisque mauris blandit. Maecenas fringilla at nulla consequat aliquam. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere [cubilia curae](https://news.ycombinator.com/).

<details open>
    <summary>Table of Contents</summary>

- [Morbi quis diam](#morbi-quis-diam)
- [Sem vitae venenatis](#sem-vitae-venenatis)
- [Nulla ultrices](#nulla-ultrices)
- [Aenean quis enim massa](#aenean-quis-enim-massa)
- [Feugiat magna porta](#feugiat-magna-porta)
</details>


## Morbi quis diam

Donec commodo urna eu diam tristique, ut faucibus sapien hendrerit. Mauris semper risus sed nibh eleifend, et feugiat nibh cursus. Nullam dapibus justo a magna eleifend, at tristique eros blandit. Morbi quis diam pellentesque, auctor tortor ut, viverra elit. Ut eu dignissim quam. Nullam scelerisque tellus facilisis, vulputate lectus id, pretium turpis. In iaculis lorem suscipit dolor tincidunt porta.


### Sed imperdiet nunc

Sed imperdiet nunc sed felis pharetra facilisis. Vestibulum consectetur eu arcu vel gravida. Fusce fringilla, ipsum eu aliquet dignissim, diam augue laoreet magna, a mollis diam nunc et elit.

#### Curabitur ornare

Curabitur ornare, sem vitae venenatis ullamcorper, ante nisl hendrerit leo, et ullamcorper justo leo et massa. Pellentesque non felis scelerisque, rutrum tellus in, tempor nulla. Nulla ultrices, nibh porttitor congue consectetur, libero purus fringilla magna

Nullam elementum, lectus quis pellentesque placerat, sapien dui dictum enim, feugiat faucibus sem sapien eget felis. Praesent non metus nunc. Phasellus non lacus sapien. Nam id lacus nec sapien dapibus finibus. Sed sed iaculis massa.

#### Nullam pulvinar velit eleifend

Nullam pulvinar velit eleifend porta fringilla. Aliquam tempus neque sit amet lobortis consequat. Curabitur quam nisl, lobortis at tortor at, rhoncus consequat nibh.

![something](/media/golang-error-handling.webp)

### Integer ultrices velit

Integer ultrices velit sit amet molestie ultricies. Integer condimentum aliquet auctor. Duis eleifend elit in lectus mollis tincidunt. Pellentesque erat ipsum, tempus ut posuere eget, hendrerit a nulla. Aliquam tempor nulla vel turpis maximus consequat.

Pellentesque ut quam viverra, auctor <mark>turpis ut, euismod dui</mark>. Phasellus luctus ullamcorper diam quis feugiat. Fusce urna dolor, sagittis et cursus in, sagittis sed nibh. Nulla malesuada mi ex, non vestibulum metus pharetra at.

> Quote here.
>
> -- <cite>Benjamin Franklin</cite>


## Sem vitae venenatis

Nulla imperdiet malesuada orci, a suscipit ex luctus sit amet. Nulla facilisi. Cras a justo euismod, suscipit lectus in, blandit turpis. Suspendisse mattis quis turpis quis congue. Nunc aliquam nisl at elit convallis iaculis.

Curabitur ornare, sem vitae venenatis ullamcorper, ante nisl hendrerit leo, et ullamcorper justo leo et massa. Donec tincidunt rhoncus imperdiet. Cras vel finibus quam.

Phasellus ac dolor sed odio euismod blandit tempor et tortor. Fusce commodo libero nec nisi viverra lacinia.

Cras non dictum nisl, in imperdiet ligula.


### Nulla ultrices

Pellentesque non felis scelerisque, rutrum tellus in, tempor nulla. Nulla ultrices, nibh porttitor congue consectetur, libero purus fringilla magna, non viverra nisi sapien a orci.

Praesent sed sem eu ante tempus mattis sed in urna. Quisque dictum dapibus diam et tempus.

- Rhoncus est in placerat
- Proin sed odio vitae metus egestas
- Sollicitudin eget quis dui
    - Donec at consequat arcu
    - vel aliquet eros. Aenean a odio enim
    - Praesent nec [tortor](https://github.com) quam.

Phasellus auctor non velit quis lacinia. Curabitur pharetra, massa vitae tristique mattis, urna risus sodales libero, vel faucibus nunc leo a erat.

Nulla facilisi. Aliquam luctus tortor turpis, non accumsan velit lacinia vitae.


## Star Wars Mixins Edition

<img src="https://media.giphy.com/media/2tDQZuljhwHTi/giphy.gif">

To illustrate another use of mixins, lets look at a Star Wars example for creating different types of Jedi and Sith instances.

We have our ForceUser class which describes someone who is force sensitive. Then we define a couple of mixins; the `JediMixin` and `SithMixin` both define custom lightsaber color and the `LightsaberMixin` uses the predefined lightsaber colors to create a lightsaber for our force user.

Finally we bring everything together by defining our `JediMaster` and `SithLord` classes. These two classes override some `ForceUser` methods to add their own custom logic.

```python
import random

class ForceUser:
    def __init__(self, name):
        self.moves = ['force push']
        self.name = name

    def get_ability(self):
        return random.choice(self.moves)

    def ability(self):
        print(f'Attacks with {self.get_ability()}')

    def __repr__(self):
        return f'I am {self.name}, a {self.__class__.__name__}!'
```


## Aenean quis enim massa

Aenean quis enim massa. Donec vel turpis et lectus dictum laoreet. Nullam eleifend vel felis id pharetra. Sed mi ipsum, malesuada non leo ut, feugiat rhoncus elit. Nullam consectetur semper quam, aliquet ullamcorper leo condimentum ac.

Nullam elementum, lectus quis pellentesque placerat, sapien dui dictum enim, feugiat faucibus sem sapien eget felis. Praesent non metus nunc. Phasellus non lacus sapien.

Nam id lacus nec sapien dapibus finibus. Sed sed iaculis massa. Sed rutrum gravida dolor quis tristique. Fusce purus magna, congue ut ligula non, euismod luctus purus. Sed id ultricies ipsum.


## Feugiat magna porta

Quisque tristique egestas nunc, sed porta magna mollis at. Phasellus feugiat pellentesque enim nec molestie. Cras pretium eu lectus et placerat. Aenean malesuada turpis quis massa ullamcorper vestibulum. Vivamus feugiat vehicula elit, quis egestas ex hendrerit ac.

Phasellus felis metus, iaculis in aliquet eu, gravida nec turpis. Cras consequat dolor justo, a feugiat magna porta a. Morbi molestie metus ac turpis pellentesque pulvinar. Nullam id suscipit libero.

Etiam et enim euismod, interdum dui ac, sodales tellus. Aenean a pellentesque neque. In ex eros, porttitor id semper molestie, commodo a dolor.


## Quam viverra

<img width="800" height="450" src="/media/food.webp" alt="Organic Food"/>
<!-- <img width="800" height="450" src="" alt=""/> -->

Cras lorem purus, ullamcorper ac lobortis sit amet, ullamcorper in augue. Vivamus posuere faucibus augue, consectetur eleifend mi mollis et. In hac habitasse platea dictumst. Ut scelerisque nulla nec tortor ultrices porta.

Nullam pulvinar velit eleifend porta fringilla. Aliquam tempus neque sit amet lobortis consequat. Curabitur quam nisl, lobortis at tortor at, rhoncus consequat nibh. Pellentesque ut quam viverra, auctor turpis ut, euismod dui. Phasellus luctus ullamcorper diam quis feugiat. Fusce urna dolor, sagittis et cursus in, sagittis sed nibh. Nulla malesuada mi ex, non vestibulum metus pharetra at.

Fusce porttitor semper augue, at porta tortor imperdiet sit amet. Sed gravida porta pellentesque.

Proin facilisis nisi molestie enim bibendum tincidunt.


## Vestibulum ante

Curabitur pharetra, massa vitae tristique mattis, urna risus sodales libero, vel faucibus nunc leo a erat. Nulla facilisi. Aliquam luctus tortor turpis, non accumsan velit lacinia vitae. Ut non ante felis. Aenean ut tempus dui, quis feugiat justo.

Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Nam sodales nibh ultrices mauris posuere, eget consequat sapien rutrum. Curabitur id ultrices enim.

### Etiam sollicitudin, quam a aliquet sodales

Vivamus in tellus hendrerit, blandit turpis sed, bibendum orci.

#### Justo dolor efficitur velit, eu tristique odio ligula quis eros

Etiam sollicitudin, quam a aliquet sodales, justo dolor efficitur velit, eu tristique odio ligula quis eros. Nunc gravida metus tempor, faucibus nisl et, rutrum purus. Suspendisse condimentum, dui et molestie ultrices, leo lectus pharetra ipsum, eu lacinia elit elit sed ante.

Sed vel lobortis orci. Sed id dignissim odio, sed viverra tortor. In aliquam cursus tortor, vel egestas nunc dictum et. Vivamus id commodo quam, quis aliquet ipsum. Etiam lobortis non ipsum ac vestibulum.


## Praesent aliquet nec

Maecenas et convallis ligula. Cras vestibulum dictum pulvinar. Phasellus ante diam, porttitor non ipsum venenatis, maximus lacinia leo. Nunc euismod ante eget massa sagittis sodales. Proin molestie maximus tempus. Sed aliquam lectus non lacus molestie lacinia.

Vivamus in tellus hendrerit, blandit turpis sed, bibendum orci. Sed ultricies viverra luctus.

Curabitur libero elit, lobortis a quam sit amet, molestie molestie urna. Etiam id porttitor elit, sed auctor quam. Curabitur dui enim, luctus nec lorem et, convallis ullamcorper risus. Praesent aliquet nec dui id tincidunt. Maecenas at est at dolor aliquam pulvinar non ut odio.

Integer sed placerat erat, vestibulum venenatis nunc.

Sed ultricies sem viverra enim egestas pulvinar. Integer id ullamcorper sapien. Quisque efficitur elit justo, eget tempus sem tincidunt ac. Morbi odio tellus, suscipit eu massa ac, tincidunt elementum felis. Etiam vel ligula a felis feugiat blandit.

Sed quis elementum quam, at pretium ex. Fusce faucibus ut lorem et tristique.


## Vitae vehicula sed

Donec nisl metus, porttitor vitae vehicula sed, tincidunt eget massa. Vivamus dapibus ante turpis, non imperdiet quam rutrum sed.

Vivamus mattis tempus ligula eu malesuada. Proin ipsum felis, varius ac rutrum et, fermentum sed nunc. Vestibulum pharetra porttitor enim id sagittis.

Curabitur bibendum, sem vel lobortis fermentum, augue ante auctor risus, molestie aliquam nunc metus ut dolor. Sed ac porttitor ante, eu aliquet leo.

Morbi in mi tincidunt, tincidunt dolor id, scelerisque justo. Integer scelerisque metus nec magna pharetra vehicula. Mauris eu tempus eros, eu ultrices ex. Nullam et ultrices dui. Nam venenatis egestas semper.

Morbi ut enim sollicitudin, pulvinar leo a, auctor eros. Interdum et malesuada fames ac ante ipsum primis in faucibus.

Nulla elit ante, lacinia sit amet lorem in, accumsan hendrerit orci. Curabitur rhoncus nunc nec scelerisque rutrum.
