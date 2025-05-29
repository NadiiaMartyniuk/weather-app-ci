package main

import (
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "net/url"
    "os"
    "time"
)

const defaultPort = "8080"

type WeatherData struct {
    Main struct {
        Temp     float64 `json:"temp"`
        Pressure float64 `json:"pressure"`
        Humidity float64 `json:"humidity"`
    } `json:"main"`
    Weather []struct {
        Description string `json:"description"`
    } `json:"weather"`
}

// Szablon HTML formularza wyboru lokalizacji z dynamicznym listami
var formTemplate = template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Pogoda</title>
</head>
<body>
  <h2>Wybierz lokalizację</h2>
  <form action="/weather" method="post">
    Kraj:
    <select id="country" name="country" onchange="updateCities()">
      <option value="PL">Polska</option>
      <option value="US">USA</option>
    </select>
    <br><br>
    Miasto:
    <select id="city" name="city">
      <!-- miasta domyślne (Polska) -->
      <option>Warszawa</option>
      <option>Kraków</option>
    </select>
    <br><br>
    <input type="submit" value="Pokaż pogodę">
  </form>

  <script>
    const citiesByCountry = {
      PL: ["Warszawa", "Kraków"],
      US: ["New York", "Los Angeles"]
    };
    function updateCities() {
      const country = document.getElementById("country").value;
      const citySelect = document.getElementById("city");
      citySelect.innerHTML = "";
      for (const city of citiesByCountry[country]) {
        const opt = document.createElement("option");
        opt.value = city;
        opt.textContent = city;
        citySelect.appendChild(opt);
      }
    }
  </script>
</body>
</html>
`))


// Szablon HTML wyniku – wyświetlenie pogody
var resultTemplate = template.Must(template.New("result").Parse(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Pogoda dla {{.City}}</title></head>
<body>
  <h2>Pogoda dla {{.City}}, {{.Country}}</h2>
  <p>Temperatura: {{.Temp}} °C</p>
  <p>Ciśnienie: {{.Pressure}} hPa</p>
  <p>Wilgotność: {{.Humidity}}%</p>
  <p>Opis: {{.Description}}</p>
  <br>
  <a href="/">Powrót</a>
</body>
</html>
`))

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = defaultPort
    }
    log.Printf("Uruchomiono aplikację: %s, Autor: Nadiia Martyniuk, Port: %s", time.Now().Format(time.RFC3339), "Nadiia Martyniuk", port)

    // Obsługa endpointu "/": wyświetlenie formularza wyboru
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if err := formTemplate.Execute(w, nil); err != nil {
            http.Error(w, "Błąd serwera", http.StatusInternalServerError)
        }
    })

    // Endpoint "/weather": odbiera dane z formularza, pobiera pogodę z OpenWeatherMap i wyświetla wynik
    http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }
        country := r.FormValue("country")
        city := r.FormValue("city")
        // Budujemy zapytanie do API pogody (w jednostkach metrycznych)
        apiKey := "adea5813315c0cc2d19df4068cab6484"
        // zakodowanie parametru q: "Miasto,Kod"
        query := url.QueryEscape(fmt.Sprintf("%s,%s", city, country))
        apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",query, apiKey,)
        resp, err := http.Get(apiURL)

        if err != nil || resp.StatusCode != 200 {
            http.Error(w, "Nie udało się pobrać danych pogodowych", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        // Parsujemy odpowiedź JSON z serwisu pogodowego
        var data WeatherData
        if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
            http.Error(w, "Błąd dekodowania danych pogodowych", http.StatusInternalServerError)
            return
        }
        // Przygotowujemy dane do szablonu i renderujemy stronę
        info := struct {
            City, Country                       string
            Temp, Pressure, Humidity           float64
            Description                         string
        }{
            City:        city,
            Country:     country,
            Temp:        data.Main.Temp,
            Pressure:    data.Main.Pressure,
            Humidity:    data.Main.Humidity,
            Description: data.Weather[0].Description,
        }
        if err := resultTemplate.Execute(w, info); err != nil {
            http.Error(w, "Błąd serwera", http.StatusInternalServerError)
        }
    })

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    log.Fatal(http.ListenAndServe(":"+port, nil))
}
