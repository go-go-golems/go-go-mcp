// Netflix Clone - Main Application with Enhanced Logging
console.log('Starting Netflix Clone with enhanced logging...');

// Register the main page with enhanced logging
registerHandler('GET', '/', (req) => {
  console.log('=== HOME PAGE REQUEST ===');
  console.log('Request object:', typeof req, req ? 'exists' : 'null/undefined');
  
  try {
    console.log('Querying featured movies...');
    const featuredMovies = db.query('SELECT * FROM movies WHERE featured = 1 ORDER BY rating DESC LIMIT 3');
    console.log('Featured movies found:', featuredMovies.length);
    
    console.log('Querying all movies...');
    const allMovies = db.query('SELECT * FROM movies ORDER BY rating DESC');
    console.log('All movies found:', allMovies.length);
    
    const totalMovieCount = db.query('SELECT COUNT(*) as count FROM movies')[0].count;
    console.log('Total movies in database:', totalMovieCount);
    
    console.log('Building HTML response...');
    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NetflixClone</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Helvetica Neue', Arial, sans-serif;
            background-color: #141414;
            color: white;
            overflow-x: hidden;
        }
        
        .navbar {
            position: fixed;
            top: 0;
            width: 100%;
            background: linear-gradient(180deg, rgba(0,0,0,0.7) 10%, transparent);
            z-index: 1000;
            padding: 20px 60px;
            transition: background-color 0.4s;
        }
        
        .navbar.scrolled {
            background-color: #141414;
        }
        
        .nav-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo {
            font-size: 32px;
            font-weight: bold;
            color: #e50914;
            text-decoration: none;
        }
        
        .nav-links {
            display: flex;
            gap: 30px;
        }
        
        .nav-links a {
            color: white;
            text-decoration: none;
            font-size: 14px;
            transition: color 0.3s;
        }
        
        .nav-links a:hover {
            color: #b3b3b3;
        }
        
        .nav-links a.active {
            color: #e50914;
            font-weight: bold;
        }
        
        .hero {
            height: 100vh;
            background: linear-gradient(rgba(0,0,0,0.4), rgba(0,0,0,0.4)), url('https://images.unsplash.com/photo-1489599735734-79b4169c2a78?w=1920&h=1080&fit=crop');
            background-size: cover;
            background-position: center;
            display: flex;
            align-items: center;
            padding: 0 60px;
        }
        
        .hero-content {
            max-width: 500px;
        }
        
        .hero-title {
            font-size: 64px;
            font-weight: bold;
            margin-bottom: 20px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.8);
        }
        
        .hero-description {
            font-size: 18px;
            line-height: 1.5;
            margin-bottom: 30px;
            text-shadow: 1px 1px 2px rgba(0,0,0,0.8);
        }
        
        .hero-buttons {
            display: flex;
            gap: 15px;
        }
        
        .btn {
            padding: 12px 30px;
            border: none;
            border-radius: 4px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: all 0.3s;
            text-decoration: none;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }
        
        .btn-primary {
            background-color: white;
            color: black;
        }
        
        .btn-primary:hover {
            background-color: rgba(255,255,255,0.8);
        }
        
        .btn-secondary {
            background-color: rgba(109,109,110,0.7);
            color: white;
        }
        
        .btn-secondary:hover {
            background-color: rgba(109,109,110,0.4);
        }
        
        .btn-favorite {
            background-color: transparent;
            color: white;
            border: 2px solid white;
            padding: 8px 16px;
            font-size: 14px;
        }
        
        .btn-favorite:hover {
            background-color: white;
            color: black;
        }
        
        .btn-favorite.favorited {
            background-color: #e50914;
            border-color: #e50914;
            color: white;
        }
        
        .btn-favorite.favorited:hover {
            background-color: #b8070f;
            border-color: #b8070f;
        }
        
        .content {
            padding: 60px;
        }
        
        .section-title {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 20px;
            color: white;
        }
        
        .movie-row {
            margin-bottom: 50px;
        }
        
        .movie-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            gap: 15px;
            overflow-x: auto;
            padding-bottom: 10px;
        }
        
        .movie-card {
            position: relative;
            border-radius: 8px;
            overflow: hidden;
            transition: transform 0.3s, box-shadow 0.3s;
            cursor: pointer;
            background: #2a2a2a;
        }
        
        .movie-card:hover {
            transform: scale(1.05);
            box-shadow: 0 8px 25px rgba(0,0,0,0.5);
        }
        
        .movie-card:hover .movie-actions {
            opacity: 1;
        }
        
        .movie-poster {
            width: 100%;
            height: 300px;
            object-fit: cover;
        }
        
        .movie-actions {
            position: absolute;
            top: 10px;
            right: 10px;
            opacity: 0;
            transition: opacity 0.3s;
        }
        
        .favorite-btn {
            background: rgba(0,0,0,0.8);
            border: none;
            color: white;
            padding: 8px;
            border-radius: 50%;
            cursor: pointer;
            font-size: 16px;
            width: 36px;
            height: 36px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s;
        }
        
        .favorite-btn:hover {
            background: rgba(0,0,0,1);
            transform: scale(1.1);
        }
        
        .favorite-btn.favorited {
            color: #e50914;
        }
        
        .movie-info {
            padding: 15px;
        }
        
        .movie-title {
            font-size: 16px;
            font-weight: bold;
            margin-bottom: 5px;
            color: white;
        }
        
        .movie-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }
        
        .movie-year {
            font-size: 12px;
            color: #b3b3b3;
        }
        
        .movie-rating {
            font-size: 12px;
            color: #46d369;
            font-weight: bold;
        }
        
        .movie-genre {
            font-size: 12px;
            color: #b3b3b3;
            margin-bottom: 8px;
        }
        
        .movie-description {
            font-size: 12px;
            color: #b3b3b3;
            line-height: 1.4;
            display: -webkit-box;
            -webkit-line-clamp: 3;
            -webkit-box-orient: vertical;
            overflow: hidden;
        }
        
        .search-container {
            margin-bottom: 30px;
        }
        
        .search-input {
            width: 100%;
            max-width: 400px;
            padding: 12px 20px;
            border: none;
            border-radius: 4px;
            background-color: rgba(255,255,255,0.1);
            color: white;
            font-size: 16px;
            backdrop-filter: blur(10px);
        }
        
        .search-input::placeholder {
            color: #b3b3b3;
        }
        
        .search-input:focus {
            outline: none;
            background-color: rgba(255,255,255,0.2);
        }
        
        .modal {
            display: none;
            position: fixed;
            z-index: 2000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.8);
            backdrop-filter: blur(5px);
        }
        
        .modal-content {
            position: relative;
            background-color: #181818;
            margin: 5% auto;
            padding: 0;
            width: 90%;
            max-width: 800px;
            border-radius: 8px;
            overflow: hidden;
        }
        
        .modal-header {
            position: relative;
            height: 400px;
            background-size: cover;
            background-position: center;
            display: flex;
            align-items: flex-end;
            padding: 40px;
        }
        
        .modal-header::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: linear-gradient(transparent, rgba(24,24,24,0.8));
        }
        
        .modal-header-content {
            position: relative;
            z-index: 1;
        }
        
        .modal-title {
            font-size: 48px;
            font-weight: bold;
            margin-bottom: 10px;
        }
        
        .modal-body {
            padding: 40px;
        }
        
        .close {
            position: absolute;
            top: 20px;
            right: 30px;
            color: white;
            font-size: 40px;
            font-weight: bold;
            cursor: pointer;
            z-index: 2;
        }
        
        .close:hover {
            color: #b3b3b3;
        }
        
        .debug-info {
            position: fixed;
            top: 10px;
            right: 10px;
            background: rgba(0,0,0,0.8);
            color: #00ff00;
            padding: 10px;
            font-family: monospace;
            font-size: 12px;
            border-radius: 4px;
            z-index: 9999;
        }
        
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #b3b3b3;
        }
        
        .empty-state h2 {
            font-size: 24px;
            margin-bottom: 16px;
            color: white;
        }
        
        .empty-state p {
            font-size: 16px;
            line-height: 1.5;
        }
        
        @media (max-width: 768px) {
            .navbar {
                padding: 15px 20px;
            }
            
            .hero {
                padding: 0 20px;
            }
            
            .hero-title {
                font-size: 40px;
            }
            
            .content {
                padding: 30px 20px;
            }
            
            .movie-grid {
                grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
            }
        }
    </style>
</head>
<body>
    <div class="debug-info">
        Featured: ${featuredMovies.length} | All: ${allMovies.length} | DB: ${totalMovieCount}
    </div>

    <nav class="navbar" id="navbar">
        <div class="nav-content">
            <a href="/" class="logo">NETFLIXCLONE</a>
            <div class="nav-links">
                <a href="/" class="active">Home</a>
                <a href="/favorites">My Favorites</a>
                <a href="#movies">Movies</a>
                <a href="#series">TV Shows</a>
                <a href="#trending">Trending</a>
            </div>
        </div>
    </nav>

    <div class="hero">
        <div class="hero-content">
            <h1 class="hero-title">${featuredMovies[0]?.title || 'Welcome'}</h1>
            <p class="hero-description">
                ${featuredMovies[0]?.description || 'Discover amazing movies and TV shows'}
            </p>
            <div class="hero-buttons">
                <button class="btn btn-primary" onclick="playMovie(${featuredMovies[0]?.id || 1})">
                    ▶ Play
                </button>
                <button class="btn btn-secondary" onclick="showMovieDetails(${featuredMovies[0]?.id || 1})">
                    ℹ More Info
                </button>
            </div>
        </div>
    </div>

    <div class="content">
        <div class="search-container">
            <input type="text" class="search-input" placeholder="Search movies and TV shows..." id="searchInput">
        </div>

        <div class="movie-row">
            <h2 class="section-title">Featured Content</h2>
            <div class="movie-grid" id="featuredGrid">
                ${featuredMovies.map(movie => `
                    <div class="movie-card" onclick="showMovieDetails(${movie.id})">
                        <img src="${movie.poster_url}" alt="${movie.title}" class="movie-poster">
                        <div class="movie-actions">
                            <button class="favorite-btn" onclick="event.stopPropagation(); toggleFavorite(${movie.id}, this)" data-movie-id="${movie.id}">
                                ♡
                            </button>
                        </div>
                        <div class="movie-info">
                            <div class="movie-title">${movie.title}</div>
                            <div class="movie-meta">
                                <span class="movie-year">${movie.year}</span>
                                <span class="movie-rating">★ ${movie.rating}</span>
                            </div>
                            <div class="movie-genre">${movie.genre}</div>
                            <div class="movie-description">${movie.description}</div>
                        </div>
                    </div>
                `).join('')}
            </div>
        </div>

        <div class="movie-row">
            <h2 class="section-title">All Movies & Shows</h2>
            <div class="movie-grid" id="allMoviesGrid">
                ${allMovies.map(movie => `
                    <div class="movie-card" onclick="showMovieDetails(${movie.id})">
                        <img src="${movie.poster_url}" alt="${movie.title}" class="movie-poster">
                        <div class="movie-actions">
                            <button class="favorite-btn" onclick="event.stopPropagation(); toggleFavorite(${movie.id}, this)" data-movie-id="${movie.id}">
                                ♡
                            </button>
                        </div>
                        <div class="movie-info">
                            <div class="movie-title">${movie.title}</div>
                            <div class="movie-meta">
                                <span class="movie-year">${movie.year}</span>
                                <span class="movie-rating">★ ${movie.rating}</span>
                            </div>
                            <div class="movie-genre">${movie.genre}</div>
                            <div class="movie-description">${movie.description}</div>
                        </div>
                    </div>
                `).join('')}
            </div>
        </div>
    </div>

    <!-- Movie Details Modal -->
    <div id="movieModal" class="modal">
        <div class="modal-content">
            <span class="close" onclick="closeModal()">&times;</span>
            <div class="modal-header" id="modalHeader">
                <div class="modal-header-content">
                    <h2 class="modal-title" id="modalTitle"></h2>
                </div>
            </div>
            <div class="modal-body">
                <div id="modalContent"></div>
            </div>
        </div>
    </div>

    <script>
        console.log('Netflix Clone JavaScript loaded');
        
        // Load favorites on page load
        let userFavorites = new Set();
        loadFavorites();
        
        // Navbar scroll effect
        window.addEventListener('scroll', function() {
            const navbar = document.getElementById('navbar');
            if (window.scrollY > 100) {
                navbar.classList.add('scrolled');
            } else {
                navbar.classList.remove('scrolled');
            }
        });

        // Search functionality
        document.getElementById('searchInput').addEventListener('input', function(e) {
            const searchTerm = e.target.value.toLowerCase();
            const movieCards = document.querySelectorAll('.movie-card');
            
            movieCards.forEach(card => {
                const title = card.querySelector('.movie-title').textContent.toLowerCase();
                const genre = card.querySelector('.movie-genre').textContent.toLowerCase();
                
                if (title.includes(searchTerm) || genre.includes(searchTerm)) {
                    card.style.display = 'block';
                } else {
                    card.style.display = 'none';
                }
            });
        });

        // Load user favorites
        function loadFavorites() {
            fetch('/api/favorites')
                .then(response => response.json())
                .then(favorites => {
                    userFavorites.clear();
                    favorites.forEach(fav => userFavorites.add(fav.movie_id));
                    updateFavoriteButtons();
                })
                .catch(error => console.error('Error loading favorites:', error));
        }

        // Update favorite button states
        function updateFavoriteButtons() {
            document.querySelectorAll('.favorite-btn').forEach(btn => {
                const movieId = parseInt(btn.dataset.movieId);
                if (userFavorites.has(movieId)) {
                    btn.classList.add('favorited');
                    btn.innerHTML = '♥';
                } else {
                    btn.classList.remove('favorited');
                    btn.innerHTML = '♡';
                }
            });
        }

        // Toggle favorite status
        function toggleFavorite(movieId, button) {
            const isFavorited = userFavorites.has(movieId);
            const action = isFavorited ? 'remove' : 'add';
            
            fetch('/api/favorites/' + action, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ movieId: movieId })
            })
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    if (action === 'add') {
                        userFavorites.add(movieId);
                        button.classList.add('favorited');
                        button.innerHTML = '♥';
                    } else {
                        userFavorites.delete(movieId);
                        button.classList.remove('favorited');
                        button.innerHTML = '♡';
                    }
                    console.log('Movie ' + movieId + (action === 'add' ? ' added' : ' removed') + ' from favorites');
                } else {
                    console.error('Failed to update favorite:', result.error);
                }
            })
            .catch(error => {
                console.error('Error toggling favorite:', error);
            });
        }

        // Movie details modal
        function showMovieDetails(movieId) {
            console.log('Showing details for movie ID:', movieId);
            fetch('/api/movie/' + movieId)
                .then(response => {
                    console.log('API response status:', response.status);
                    return response.json();
                })
                .then(movie => {
                    console.log('Movie data received:', movie);
                    const isFavorited = userFavorites.has(movie.id);
                    
                    document.getElementById('modalTitle').textContent = movie.title;
                    document.getElementById('modalHeader').style.backgroundImage = 'url(' + movie.poster_url + ')';
                    document.getElementById('modalContent').innerHTML = 
                        '<div style="display: flex; gap: 20px; margin-bottom: 20px;">' +
                            '<span style="color: #46d369; font-weight: bold;">★ ' + movie.rating + '</span>' +
                            '<span>' + movie.year + '</span>' +
                            '<span>' + movie.duration + ' min</span>' +
                            '<span style="background: #333; padding: 2px 8px; border-radius: 4px; font-size: 12px;">' + movie.genre + '</span>' +
                        '</div>' +
                        '<p style="line-height: 1.6; margin-bottom: 20px;">' + movie.description + '</p>' +
                        '<div style="display: flex; gap: 15px;">' +
                            '<button class="btn btn-primary" onclick="playMovie(' + movie.id + ')">▶ Play</button>' +
                            '<button class="btn btn-favorite ' + (isFavorited ? 'favorited' : '') + '" onclick="toggleFavoriteModal(' + movie.id + ', this)">' +
                                (isFavorited ? '♥ Remove from Favorites' : '♡ Add to Favorites') +
                            '</button>' +
                        '</div>';
                    document.getElementById('movieModal').style.display = 'block';
                })
                .catch(error => {
                    console.error('Error fetching movie details:', error);
                });
        }

        // Toggle favorite from modal
        function toggleFavoriteModal(movieId, button) {
            const isFavorited = userFavorites.has(movieId);
            const action = isFavorited ? 'remove' : 'add';
            
            fetch('/api/favorites/' + action, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ movieId: movieId })
            })
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    if (action === 'add') {
                        userFavorites.add(movieId);
                        button.classList.add('favorited');
                        button.innerHTML = '♥ Remove from Favorites';
                    } else {
                        userFavorites.delete(movieId);
                        button.classList.remove('favorited');
                        button.innerHTML = '♡ Add to Favorites';
                    }
                    updateFavoriteButtons(); // Update all buttons on page
                }
            })
            .catch(error => {
                console.error('Error toggling favorite:', error);
            });
        }

        function closeModal() {
            document.getElementById('movieModal').style.display = 'none';
        }

        function playMovie(movieId) {
            console.log('Play movie called for ID:', movieId);
            alert('Playing movie with ID: ' + movieId + '\nIn a real app, this would start video playback!');
        }

        // Close modal when clicking outside
        window.onclick = function(event) {
            const modal = document.getElementById('movieModal');
            if (event.target == modal) {
                closeModal();
            }
        }
    </script>
</body>
</html>
    `;
    
    console.log('Returning HTML response...');
    console.log('=== HOME PAGE REQUEST COMPLETE ===');
    
    return Response.html(html);
    
  } catch (error) {
    console.error('ERROR in home page handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

// Favorites page
registerHandler('GET', '/favorites', (req) => {
  console.log('=== FAVORITES PAGE REQUEST ===');
  
  try {
    console.log('Querying favorite movies...');
    const favoriteMovies = db.query(`
      SELECT m.* FROM movies m 
      INNER JOIN favorites f ON m.id = f.movie_id 
      ORDER BY f.added_at DESC
    `);
    console.log('Favorite movies found:', favoriteMovies.length);
    
    const totalMovieCount = db.query('SELECT COUNT(*) as count FROM movies')[0].count;
    const favoriteCount = favoriteMovies.length;
    
    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Favorites - NetflixClone</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Helvetica Neue', Arial, sans-serif;
            background-color: #141414;
            color: white;
            overflow-x: hidden;
            padding-top: 80px;
        }
        
        .navbar {
            position: fixed;
            top: 0;
            width: 100%;
            background-color: #141414;
            z-index: 1000;
            padding: 20px 60px;
        }
        
        .nav-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo {
            font-size: 32px;
            font-weight: bold;
            color: #e50914;
            text-decoration: none;
        }
        
        .nav-links {
            display: flex;
            gap: 30px;
        }
        
        .nav-links a {
            color: white;
            text-decoration: none;
            font-size: 14px;
            transition: color 0.3s;
        }
        
        .nav-links a:hover {
            color: #b3b3b3;
        }
        
        .nav-links a.active {
            color: #e50914;
            font-weight: bold;
        }
        
        .content {
            padding: 60px;
        }
        
        .page-header {
            margin-bottom: 40px;
        }
        
        .page-title {
            font-size: 48px;
            font-weight: bold;
            margin-bottom: 16px;
            color: white;
        }
        
        .page-subtitle {
            font-size: 18px;
            color: #b3b3b3;
            margin-bottom: 20px;
        }
        
        .stats {
            display: flex;
            gap: 30px;
            margin-bottom: 40px;
        }
        
        .stat-item {
            background: rgba(255,255,255,0.1);
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }
        
        .stat-number {
            font-size: 32px;
            font-weight: bold;
            color: #e50914;
            display: block;
        }
        
        .stat-label {
            font-size: 14px;
            color: #b3b3b3;
            margin-top: 8px;
        }
        
        .movie-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
            gap: 15px;
            overflow-x: auto;
            padding-bottom: 10px;
        }
        
        .movie-card {
            position: relative;
            border-radius: 8px;
            overflow: hidden;
            transition: transform 0.3s, box-shadow 0.3s;
            cursor: pointer;
            background: #2a2a2a;
        }
        
        .movie-card:hover {
            transform: scale(1.05);
            box-shadow: 0 8px 25px rgba(0,0,0,0.5);
        }
        
        .movie-card:hover .movie-actions {
            opacity: 1;
        }
        
        .movie-poster {
            width: 100%;
            height: 300px;
            object-fit: cover;
        }
        
        .movie-actions {
            position: absolute;
            top: 10px;
            right: 10px;
            opacity: 0;
            transition: opacity 0.3s;
        }
        
        .favorite-btn {
            background: rgba(0,0,0,0.8);
            border: none;
            color: #e50914;
            padding: 8px;
            border-radius: 50%;
            cursor: pointer;
            font-size: 16px;
            width: 36px;
            height: 36px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s;
        }
        
        .favorite-btn:hover {
            background: rgba(0,0,0,1);
            transform: scale(1.1);
        }
        
        .movie-info {
            padding: 15px;
        }
        
        .movie-title {
            font-size: 16px;
            font-weight: bold;
            margin-bottom: 5px;
            color: white;
        }
        
        .movie-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }
        
        .movie-year {
            font-size: 12px;
            color: #b3b3b3;
        }
        
        .movie-rating {
            font-size: 12px;
            color: #46d369;
            font-weight: bold;
        }
        
        .movie-genre {
            font-size: 12px;
            color: #b3b3b3;
            margin-bottom: 8px;
        }
        
        .movie-description {
            font-size: 12px;
            color: #b3b3b3;
            line-height: 1.4;
            display: -webkit-box;
            -webkit-line-clamp: 3;
            -webkit-box-orient: vertical;
            overflow: hidden;
        }
        
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #b3b3b3;
        }
        
        .empty-state h2 {
            font-size: 24px;
            margin-bottom: 16px;
            color: white;
        }
        
        .empty-state p {
            font-size: 16px;
            line-height: 1.5;
            margin-bottom: 20px;
        }
        
        .btn {
            padding: 12px 30px;
            border: none;
            border-radius: 4px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: all 0.3s;
            text-decoration: none;
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }
        
        .btn-primary {
            background-color: #e50914;
            color: white;
        }
        
        .btn-primary:hover {
            background-color: #b8070f;
        }
        
        .debug-info {
            position: fixed;
            top: 10px;
            right: 10px;
            background: rgba(0,0,0,0.8);
            color: #00ff00;
            padding: 10px;
            font-family: monospace;
            font-size: 12px;
            border-radius: 4px;
            z-index: 9999;
        }
        
        @media (max-width: 768px) {
            .navbar {
                padding: 15px 20px;
            }
            
            .content {
                padding: 30px 20px;
            }
            
            .page-title {
                font-size: 32px;
            }
            
            .stats {
                flex-direction: column;
                gap: 15px;
            }
            
            .movie-grid {
                grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
            }
        }
    </style>
</head>
<body>
    <div class="debug-info">
        Favorites: ${favoriteCount} | Total: ${totalMovieCount}
    </div>

    <nav class="navbar">
        <div class="nav-content">
            <a href="/" class="logo">NETFLIXCLONE</a>
            <div class="nav-links">
                <a href="/">Home</a>
                <a href="/favorites" class="active">My Favorites</a>
                <a href="#movies">Movies</a>
                <a href="#series">TV Shows</a>
                <a href="#trending">Trending</a>
            </div>
        </div>
    </nav>

    <div class="content">
        <div class="page-header">
            <h1 class="page-title">My Favorites</h1>
            <p class="page-subtitle">Your personally curated collection of amazing content</p>
        </div>

        <div class="stats">
            <div class="stat-item">
                <span class="stat-number">${favoriteCount}</span>
                <div class="stat-label">Favorite Movies</div>
            </div>
            <div class="stat-item">
                <span class="stat-number">${favoriteMovies.length > 0 ? Math.round(favoriteMovies.reduce((sum, movie) => sum + movie.rating, 0) / favoriteMovies.length * 10) / 10 : 0}</span>
                <div class="stat-label">Average Rating</div>
            </div>
        </div>

        ${favoriteMovies.length > 0 ? `
            <div class="movie-grid">
                ${favoriteMovies.map(movie => `
                    <div class="movie-card" onclick="showMovieDetails(${movie.id})">
                        <img src="${movie.poster_url}" alt="${movie.title}" class="movie-poster">
                        <div class="movie-actions">
                            <button class="favorite-btn" onclick="event.stopPropagation(); removeFavorite(${movie.id}, this)" data-movie-id="${movie.id}">
                                ♥
                            </button>
                        </div>
                        <div class="movie-info">
                            <div class="movie-title">${movie.title}</div>
                            <div class="movie-meta">
                                <span class="movie-year">${movie.year}</span>
                                <span class="movie-rating">★ ${movie.rating}</span>
                            </div>
                            <div class="movie-genre">${movie.genre}</div>
                            <div class="movie-description">${movie.description}</div>
                        </div>
                    </div>
                `).join('')}
            </div>
        ` : `
            <div class="empty-state">
                <h2>No favorites yet</h2>
                <p>Start building your collection by adding movies to your favorites.<br>
                Browse our catalog and click the heart icon on any movie you love!</p>
                <a href="/" class="btn btn-primary">Browse Movies</a>
            </div>
        `}
    </div>

    <script>
        console.log('Favorites page loaded');

        function removeFavorite(movieId, button) {
            if (confirm('Remove this movie from your favorites?')) {
                fetch('/api/favorites/remove', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ movieId: movieId })
                })
                .then(response => response.json())
                .then(result => {
                    if (result.success) {
                        // Remove the movie card from the page
                        button.closest('.movie-card').remove();
                        
                        // Update stats
                        const statNumber = document.querySelector('.stat-number');
                        const currentCount = parseInt(statNumber.textContent);
                        statNumber.textContent = currentCount - 1;
                        
                        // Show empty state if no favorites left
                        if (currentCount - 1 === 0) {
                            location.reload();
                        }
                        
                        console.log('Movie ' + movieId + ' removed from favorites');
                    } else {
                        console.error('Failed to remove favorite:', result.error);
                        alert('Failed to remove from favorites. Please try again.');
                    }
                })
                .catch(error => {
                    console.error('Error removing favorite:', error);
                    alert('Failed to remove from favorites. Please try again.');
                });
            }
        }

        function showMovieDetails(movieId) {
            // Redirect to home page with movie details
            window.location.href = '/?movie=' + movieId;
        }
    </script>
</body>
</html>
    `;
    
    console.log('Returning favorites page...');
    console.log('=== FAVORITES PAGE REQUEST COMPLETE ===');
    
    return Response.html(html);
    
  } catch (error) {
    console.error('ERROR in favorites page handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

// API endpoint to get user favorites
registerHandler('GET', '/api/favorites', (req) => {
  console.log('=== FAVORITES API REQUEST ===');
  
  try {
    const favorites = db.query('SELECT movie_id, added_at FROM favorites ORDER BY added_at DESC');
    console.log('Favorites found:', favorites.length);
    
    return Response.json(favorites);
    
  } catch (error) {
    console.error('ERROR in favorites API handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

// API endpoint to add movie to favorites
registerHandler('POST', '/api/favorites/add', (req) => {
  console.log('=== ADD FAVORITE API REQUEST ===');
  
  try {
    const { movieId } = JSON.parse(req.Body || '{}');
    console.log('Adding movie to favorites:', movieId);
    
    if (!movieId) {
      return Response.error('Movie ID is required', 400);
    }
    
    // Check if movie exists
    const movie = db.query('SELECT id FROM movies WHERE id = ?', movieId)[0];
    if (!movie) {
      return Response.error('Movie not found', 404);
    }
    
    // Add to favorites (ignore if already exists due to UNIQUE constraint)
    const result = db.exec('INSERT OR IGNORE INTO favorites (movie_id) VALUES (?)', movieId);
    
    if (result.success) {
      console.log('Movie added to favorites successfully');
      return Response.json({ success: true, message: 'Added to favorites' });
    } else {
      console.error('Failed to add to favorites:', result.error);
      return Response.error('Failed to add to favorites', 500);
    }
    
  } catch (error) {
    console.error('ERROR in add favorite API handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

// API endpoint to remove movie from favorites
registerHandler('POST', '/api/favorites/remove', (req) => {
  console.log('=== REMOVE FAVORITE API REQUEST ===');
  
  try {
    const { movieId } = JSON.parse(req.Body || '{}');
    console.log('Removing movie from favorites:', movieId);
    
    if (!movieId) {
      return Response.error('Movie ID is required', 400);
    }
    
    const result = db.exec('DELETE FROM favorites WHERE movie_id = ?', movieId);
    
    if (result.success) {
      console.log('Movie removed from favorites successfully');
      return Response.json({ success: true, message: 'Removed from favorites' });
    } else {
      console.error('Failed to remove from favorites:', result.error);
      return Response.error('Failed to remove from favorites', 500);
    }
    
  } catch (error) {
    console.error('ERROR in remove favorite API handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

// API endpoint to get movie details
registerHandler('GET', '/api/movie/:id', (req) => {
  console.log('=== MOVIE API REQUEST ===');
  console.log('Request params:', req.Params);
  
  try {
    const movieId = req.Params.id;
    console.log('Looking for movie with ID:', movieId);
    
    const movie = db.query('SELECT * FROM movies WHERE id = ?', movieId)[0];
    console.log('Movie found:', movie ? movie.title : 'not found');
    
    if (!movie) {
      console.log('Movie not found, sending 404');
      return Response.error('Movie not found', 404);
    }
    
    console.log('Sending movie data');
    console.log('=== MOVIE API REQUEST COMPLETE ===');
    
    return Response.json(movie);
    
  } catch (error) {
    console.error('ERROR in movie API handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

// API endpoint to search movies
registerHandler('GET', '/api/search', (req) => {
  console.log('=== SEARCH API REQUEST ===');
  console.log('Query params:', req.Query);
  
  try {
    const query = req.Query.q || '';
    console.log('Search query:', query);
    
    const movies = db.query(
      'SELECT * FROM movies WHERE title LIKE ? OR genre LIKE ? ORDER BY rating DESC',
      `%${query}%`, `%${query}%`
    );
    console.log('Search results:', movies.length, 'movies found');
    
    console.log('=== SEARCH API REQUEST COMPLETE ===');
    
    return Response.json(movies);
    
  } catch (error) {
    console.error('ERROR in search API handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

// API endpoint to get movies by genre
registerHandler('GET', '/api/genre/:genre', (req) => {
  console.log('=== GENRE API REQUEST ===');
  console.log('Genre param:', req.Params.genre);
  
  try {
    const genre = req.Params.genre;
    console.log('Looking for genre:', genre);
    
    const movies = db.query(
      'SELECT * FROM movies WHERE genre LIKE ? ORDER BY rating DESC',
      `%${genre}%`
    );
    console.log('Genre results:', movies.length, 'movies found');
    
    console.log('=== GENRE API REQUEST COMPLETE ===');
    
    return Response.json(movies);
    
  } catch (error) {
    console.error('ERROR in genre API handler:', error);
    return Response.error('Internal Server Error: ' + error.message, 500);
  }
});

console.log('Netflix Clone with Favorites system registered successfully! Visit http://localhost:8080 to see it in action.'); 