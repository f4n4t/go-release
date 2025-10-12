package release_test

import (
	"testing"

	"github.com/f4n4t/go-release"
	"github.com/stretchr/testify/assert"
)

func TestParseSection(t *testing.T) {
	releaseService := release.NewServiceBuilder().WithSkipPre(true).Build()

	tests := []struct {
		name        string
		releaseName string
		preSection  string
		expected    release.Section
	}{
		// Movies - Standard cases
		{"Movie - German", "Die.Abenteurer.1967.German.1080p.BluRay.x264-DETAiLS", "", release.Movies},
		{"Movie - German", "I.Kill.Giants.2017.German.DL.DTSHD.1080p.BDRip.x265-sikamikanico", "", release.Movies},
		{"Movie - Complete Bluray", "Faceless.1987.COMPLETE.BLURAY-FULLBRUTALiTY", "", release.Movies},
		{"Movie - 4K Remastered", "Super.Mario.Bros.1993.4K.Remastered.Dual.Custom.AUS.UHD.BluRay-MAMA", "", release.Movies},
		{"Movie - Subbed", "Biggie.Das.ist.meine.Geschichte.2021.German.Subbed.AC3.1080p.WebRip.x265-FuN", "", release.Movies},
		{"Movie - 1080i Format", "Cold.Prey.Eiskalter.Tod.2006.German.DTS-HD.1080i.BluRay.MPEG-2.REMUX-LeetHD", "", release.Movies},
		{"Movie - HEVC", "Star.Wars.The.Rise.Of.Skywalker.2019.German.HEVC.1080p.BluRay-QfG", "", release.Movies},
		{"Movie - UHD.BluRay", "Godzilla.vs.Kong.2021.UHD.BluRay.2160p.DTS-HD.MA5.1.HEVC.REMUX-FraMeSToR", "", release.Movies},

		// Movies - Edge cases
		{"Movie - HDR Format", "The.Batman.2022.HDR.2160p.WEB.H265-EMPATHY", "", release.Movies},
		{"Movie - DV Format", "Dune.2021.DV.2160p.WEB.H265-TIMECUT", "", release.Movies},
		{"Movie - No Year", "The.Matrix.REMASTERED.1080p.BluRay.x264-LEVERAGE", "", release.Movies},
		{"Movie - With Dots Only", "Star.Wars.Episode.IV.A.New.Hope.1977.1080p.BluRay.DTS.x264-CtrlHD", "", release.Movies},
		{"Movie - Mixed Separators", "The_Lord.of-the_Rings.2001.Extended.1080p.BluRay.x264-GROUP", "", release.Movies},
		{"Movie - With Underscores", "Mad_Max_Fury_Road_2015_1080p_BluRay_x264_DTS-JYK", "", release.Movies},

		// TV Shows - Standard episodes
		{"TV - German Documentary", "Spiegel.TV.2022-07-11.GERMAN.DOKU.1080p.WEB.x264-TSCC", "", release.TV},
		{"TV - Single Episode", "The.Last.of.Us.S01E03.1080p.WEB.H264-CAKES", "", release.TV},
		{"TV - Reality Show", "The.Great.British.Bake.Off.S13E06.1080p.WEB.H264-GOSSIP", "", release.TV},
		{"TV - Show with Date", "Late.Night.2023.05.25.Harrison.Ford.1080p.WEB.h264-BAE", "", release.TV},
		{"TV - Daily Show", "The.Daily.Show.2023.08.15.1080p.WEB.h264-WHOSNEXT", "", release.TV},
		{"TV - Complete Bluray", "Westworld.S04D01.COMPLETE.BLURAY-BROADCAST", "", release.TV},
		{"TV - Multi-language Pack", "Mr.Robot.S03.MULTi.COMPLETE.BLURAY-SharpHD", "", release.TVPack},

		// TV Shows - Season packs and alternative formats
		{"TV - Anime Season", "To.Your.Eternity.2021.S01.ANiME.German.AAC.1080p.WEBRiP.HEVC-DS7", "", release.TVPack},
		{"TV - Season Pack", "Game.of.Thrones.S08.1080p.BluRay.x264-ROVERS", "", release.TVPack},
		{"TV - Season with Episodes", "Friends.S01E01-E24.1080p.BluRay.x264-GROUP", "", release.TV},
		{"TV - Alternative Episode Format", "Succession.1x09.1080p.WEB.H264-GLHF", "", release.TV},
		{"TV - Alternative Season Format", "The.Office.S05.D01.German.DL.1080p.BluRay.x264-RSG", "", release.TVPack},

		// Music - FLAC format
		{"Music - FLAC - WEB", "H.E.A.T-Freedom_Rock-2023_Version-24BIT-WEB-FLAC-2023-TiMES", "", release.AudioFLAC},
		{"Music - FLAC - SACD", "Johann_Sebastian_Bach-Complete_Organ_Works_played_on_Silbermann_Organs-SACD-FLAC-2012-TSiNT", "", release.AudioFLAC},
		{"Music - FLAC - Multiple CDs", "Pink_Floyd-The_Dark_Side_Of_The_Moon-2CD-FLAC-1973-EMG", "", release.AudioFLAC},
		{"Music - FLAC - Vinyl Rip", "Beatles-Abbey_Road-VINYL-FLAC-1969-CUSTODES", "", release.AudioFLAC},

		// Music - missing source
		{"Music - FLAC - Album", "Radiohead-In_Rainbows-FLAC-2007-EMG", "", release.Unknown},

		// Music - MP3 format with various sources
		{"Music - MP3 - WEB", "Melonboy-You_Should_Be_Here-(SR059)-WEB-2025-BB", "", release.AudioMP3},
		{"Music - MP3 - SAT", "KI_KI_-_Radio_1s_Essential_Mix-SAT-04-26-2025-TALiON", "", release.AudioMP3},
		{"Music - MP3 - DVBs", "Ben_Liebrand--In_The_House_(SSL)-DVBS-04-26-2025-OMA", "", release.AudioMP3},
		{"Music - MP3 - Cable", "Fedde_Le_Grand_and_Funkerman_-_Dance_Department_(Radio538)-CABLE-04-19-2025-TALiON", "", release.AudioMP3},
		{"Music - MP3 - CDR", "Exsanguinment-Remnants_of_Putrefaction-CDR-2024-DiTCH", "", release.AudioMP3},
		{"Music - MP3 - Single CD (CDS)", "Common-The_6th_Sense_(Something_U_Feel)-Promo_CDS-2000-GCP_INT", "", release.AudioMP3},
		{"Music - MP3 - Maxi CD (CDM)", "Common-The_6th_Sense_(Something_U_Feel)-Promo_CDM-2000-GCP_INT", "", release.AudioMP3},
		{"Music - MP3 - Album single CD", "VA-Eurovision_Song_Contest_Basel_2025-CD-2025-C4", "", release.AudioMP3},
		{"Music - MP3 - DVD", "Bonez_MC_und_Raf_Camora-Palmen_Aus_Plastik_Live_in_Stuttgart-DVD-DE-2016-NOiR", "", release.AudioMP3},
		{"Music - MP3 - Album multiple CDs", "VA-Eurovision_Song_Contest_Basel_2025-2CD-2025-C4", "", release.AudioMP3},
		{"Music - MP3 - Vinyl", "Scruscru-LTDWLBL010-(LTDWLBL010)-VINYL-2024-EMP", "", release.AudioMP3},
		{"Music - MP3 - Tape", "Old_School_Hip_Hop_Mix-TAPE-2003-CMS", "", release.AudioMP3},
		{"Music - MP3 - VLS", "Moodymann-Tribute-(KDJ48)-VLS-2016-USR", "", release.AudioMP3},

		// Music Videos
		{"AudioVideo - Standard Format", "2_Unlimited-The_Real_Thing-DVDRiP-x264-1994-ZViD_iNT", "", release.AudioVideo},
		{"AudioVideo - Concert", "Metallica-Live_In_Seattle-KONZERT-DVDRIP-x264-2008-GRMV", "", release.AudioVideo},
		{"AudioVideo - Music Video", "Beyonce-Formation-DVDRip-x264-2016-FRAY", "", release.AudioVideo},
		{"AudioVideo - MBluRay Format", "Adele-Live_At_The_Royal_Albert_Hall-MBLURAY-x264-2011-FKKHD", "", release.AudioVideo},
		{"AudioVideo - With Date", "Michael_Jackson-Thriller-MBLURAY-x264-1983-FKKFHD", "", release.AudioVideo},
		{"AudioVideo - Without Codec", "KrawallBrueder.25.Jahre.Live.2022.GERMAN.COMPLETE.MBLURAY-FULLBRUTALiTY", "", release.AudioVideo},

		// Audiobooks
		{"Audiobook - Standard Format", "Stephen_King-The_Stand-AUDIOBOOK-WEB-2020-MOO", "", release.AudioBooks},
		{"Audiobook - German", "Marc_Uwe_Kling-Die_Kaenguru_Chroniken-5CD-DE-AUDIOBOOK-PROPER-FLAC-2012-VOiCE", "", release.AudioBooks},
		{"Audiobook - MP3 Format", "J_K_Rowling-Harry_Potter_And_The_Philosophers_Stone-UNABRIDGED-WEB-2015-plixid-ABOOK", "", release.AudioBooks},
		{"Audiobook - Alternative Tag", "George_RR_Martin-A_Game_of_Thrones-HOERBUCH-CD-DE-2012-VOLDiES", "", release.AudioBooks},

		// Games - Various platforms
		{"Games - Windows", "Call.of.Duty.Modern.Warfare.III.PROPER-RAZOR1911", "games", release.GamesWindows},
		{"Games - Xbox", "Resident.Evil.6.HD.USA.XBOXONE-BigBlueBox", "", release.GamesXbox},
		{"Games - Xbox Series X", "Halo.Infinite.XBOX-RUNE", "", release.GamesXbox},
		{"Games - PlayStation", "Gran.Turismo.7.PS5-DUPLEX", "", release.GamesPlaystation},
		{"Games - PlayStation 4", "God.of.War.Ragnarok.PS4-DUPLEX", "", release.GamesPlaystation},
		{"Games - PlayStation 3", "The.Last.of.Us.PS3-DUPLEX", "", release.GamesPlaystation},
		{"Games - Nintendo Switch", "Super.Mario.Odyssey.NSW-VENOM", "", release.GamesNintendo},
		{"Games - Wii", "Mario.Kart.Wii-NRP", "", release.GamesNintendo},
		{"Games - WiiU", "Mario.Kart.8.WiiU-VENOM", "", release.GamesNintendo},
		{"Games - Linux", "Unforeseen_Incidents_v1.2_Linux-Razor1911", "", release.GamesLinux},
		{"Games - Linux with PreSection", "HalfLife.2.Linux-PLAZA", "games", release.GamesLinux},
		{"Games - MacOS with PreSection", "Cities.Skylines.MacOS-ACTiVATED", "games", release.GamesMacOS},

		// Apps
		{"Apps - Windows", "Adobe.Photoshop.2023.v24.0.0.59-RUSTED", "0day", release.AppsWindows},
		{"Apps - Windows Explicit", "Microsoft.Office.2021.Professional.Plus.Windows-CrackzSoft", "apps", release.AppsWindows},
		{"Apps - MacOS", "Final.Cut.Pro.10.6.5.macOS-TNT", "apps", release.AppsMacOS},
		{"Apps - Linux", "Autodesk.Maya.2023.Linux-AMPED", "apps", release.AppsLinux},
		{"Apps - Crossplatform", "VirtualBox.7.0.0.BETA4.Crossplatform-F4CG", "apps", release.AppsMisc},
		{"Apps - VSTi", "Native.Instruments.Massive.X.v1.3.5.VSTi.VST3.AAX.x64-R2R", "0day", release.AppsWindows},

		// Learning Materials
		{"Tutorial - Algorithm Series", "PEARSON.ALGORITHMS.24-PART.LECTURE.SERIES-iLLiTERATE", "", release.Tutorials},
		{"Tutorial - Programming Course", "Udemy.Complete.Python.Developer.Course.2023-KNOWLEDGE", "", release.Tutorials},

		// Sports
		{"Sport - Tennis Match", "Tennis.Wimbledon.2022.Frauen.Finale.Jabeur.vs.Rybakina.GERMAN.1080p.HDTV.x264-TSCC", "", release.Sport},
		{"Sport - Football Match", "Fussball.Laenderspiel.2024-06-07.Deutschland.vs.Griechenland.GERMAN.1080p.WEB.H264-TSCC", "", release.Sport},
		{"Sport - Formula 1", "Formula1.2023.Monaco.Grand.Prix.Race.1080p.HDTV.x264-VERUM", "", release.Sport},
		{"Sport - Basketball", "NBA.2023.Finals.Game7.Nuggets.vs.Heat.1080p.WEB.h264-SURFER", "", release.Sport},
		{"Sport - Football League", "Bundesliga.2023.09.23.Bayern.vs.Bochum.German.1080p.WEB.h264-WEBKiNGHD", "", release.Sport},
		{"Sport - MMA", "UFC.293.Adesanya.vs.Strickland.Main.Card.1080p.WEB.h264-VERUM", "", release.Sport},
		{"Sport - Boxing", "Boxing.2023.05.20.Taylor.vs.Cameron.1080p.HDTV.x264-VERUM", "", release.Sport},

		// E-Books
		{"Ebook - Magazine", "Tichys.Einblick.No.08.2022.GERMAN.HYBRID.MAGAZINE.eBook-LORENZ", "", release.Ebooks},
		{"Ebook - Book Series", "Jeny.Han.Sommer.Band.1-3.2011-2012.German.Retail.EPUB.eBook", "", release.Ebooks},
		{"Ebook - Novel", "Stephen.King.The.Institute.2019.Retail.ePub-BEAN", "", release.Ebooks},
		{"Ebook - PDF Format", "Computer.Architecture.6th.Edition.2022.PDF-CUSTODES", "", release.Ebooks},
		{"Ebook - Technical Manual", "Microsoft.Azure.Administrator.Study.Guide.2023.EPUB-LEARNING", "", release.Ebooks},
		{"Ebook - Encyclopedia", "Encyclopedia.Britannica.2022.Edition.EPUB-KNOWLEDGE", "", release.Ebooks},
		{"Ebook - Comic", "Amazing.Spider-Man.Volume.1.Issues.1-10.CBR-COMICS", "", release.Ebooks},
		{"Ebook - Manga", "Attack.On.Titan.Volume.1-5.2013.EPUB-MANGA", "", release.Ebooks},

		// Mobile
		{"Mobile - Android App", "Sygic.GPS.Navigation.v22.0.6.ANDROiD.CELEBRATiON-rGPDA", "", release.Mobile},
		{"Mobile - Android Game", "Minecraft.Pocket.Edition.v1.19.51.ANDROiD-rGPDA", "", release.Mobile},
		{"Mobile - Android with APK", "WhatsApp.Messenger.v2.23.5.12.Android-iND.apk", "", release.Mobile},
		{"Mobile - Android Business App", "Microsoft.Office.v16.0.15629.20208.ANDROiD-rGPDA", "", release.Mobile},
		{"Mobile - Android Banking", "Deutsche.Bank.App.v2.7.2.ANDROiD-PierDreM", "", release.Mobile},

		// XXX content - Various categories
		{"XXX - Imageset", "PlayboyPlus.22.06.15.Khloe.Terae.Poolside.Perfection.XXX.IMAGESET-YAPG", "", release.XXXImagesets},
		{"XXX - Alternative Imageset", "MetArt.22.07.10.Hazel.Moore.Hot.Summer.XXX.imagesets-ODDBALL", "", release.XXXImagesets},
		{"XXX - Clips", "FTVGirls.22.06.28.Cassidy.Toys.And.Squirting.Fountain.XXX.1080p.MP4-YAPG", "", release.XXXClips},
		{"XXX - DVDRip", "Dont.Tell.My.Wife.5.XXX.DVDRip.x264-UPPERCUT", "", release.XXXMovies},
		{"XXX - DVD Standard", "Teen.Dreams.XXX.DVD5-STARLETS", "", release.XXXDVD},
		{"XXX - DVD Alternative", "Busty.Slumber.Party.XXX.DVDR-Pr0nStarS", "", release.XXXDVD},
		{"XXX - Pack", "JacquieEtMichelTV.22.07.01-07.10.COMPLETE.XXX.PACK-YAPG", "", release.XXXPack},
		{"XXX - Movies", "Private.Gold.The.Very.Best.Of.Private.Gold.XXX.720p.WEBRip.MP4-VSEX", "", release.XXXMovies},
		{"XXX - Standard", "Evil.Angel.22.06.24.Carmela.Clutch.Anal.Dream.Girl.XXX.1080p.HEVC.x265-GLUTEuS", "", release.XXXClips},

		// Unknown and edge cases
		{"Unknown - No Recognizable Pattern", "Was.Willst.Du", "", release.Unknown},
		{"Unknown - Random Filename", "abcdefg.123456.xyz", "", release.Unknown},
		{"Unknown - Minimal Info", "MyHomeVideo", "", release.Unknown},
		{"Unknown - Numeric Only", "12345678", "", release.Unknown},

		// Complex mixed cases that test the hierarchy of detection
		{"Mixed - Has XXX but is Documentary", "XXX.The.Documentary.2002.1080p.WEB.H264-WAVES", "", release.Movies},
		{"Mixed - TV with Game Name", "Game.of.Thrones.S01E01.1080p.BluRay.x264-ROVERS", "", release.TV},
		{"Mixed - Sport Documentary", "Football.The.Hard.Way.Documentary.2022.1080p.WEB.H264-BURN", "", release.Movies},
		{"Mixed - TV About Music", "Classic.Albums.S01E05.Who.Live.at.Leeds.1080p.BluRay.X264-WASTE", "", release.TV},
		{"Mixed - Movie About Games", "Free.Guy.2021.1080p.WEB.H264-NAISU", "", release.Movies},
	}

	// Run all tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preInfo := &release.Pre{Section: tt.preSection}
			gotSection := releaseService.ParseSection(tt.releaseName, preInfo)
			assert.Equal(t, tt.expected, gotSection, "Release: %s", tt.releaseName)
		})
	}
}

func TestParseResolution(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected release.Resolution
	}{
		{"Standard SD", "Movie.Title.2023.DVDRip.x264-GROUP", release.SD},
		{"Standard HD 720p", "Movie.Title.2023.720p.BluRay.x264-GROUP", release.HD},
		{"Standard FHD 1080p", "Movie.Title.2023.1080p.BluRay.x264-GROUP", release.FHD},
		{"Standard UHD 2160p", "Movie.Title.2023.2160p.UHD.BluRay.x265-GROUP", release.UHD},

		{"Interlaced HD 720i", "Movie.Title.2023.720i.HDTV.x264-GROUP", release.HD},
		{"Interlaced FHD 1080i", "Movie.Title.2023.1080i.BluRay.x264-GROUP", release.FHD},
		{"Interlaced UHD 2160i", "Movie.Title.2023.2160i.HDTV.x264-GROUP", release.UHD},

		{"Complete BluRay", "Movie.Title.2023.COMPLETE.BLURAY-GROUP", release.FHD},
		{"UHD BluRay", "Movie.Title.2023.UHD.BLURAY-GROUP", release.UHD},
		{"FHD explicit", "Movie.Title.2023.FHD.BluRay.x264-GROUP", release.FHD},

		{"No resolution info", "Movie.Title.2023.BluRay.x264-GROUP", release.SD},
		{"Very old format", "Movie.Title.2023.XviD-GROUP", release.SD},

		{"Mixed case 720P", "Movie.Title.2023.720P.BluRay.x264-GROUP", release.HD},
		{"Mixed case 1080P", "Movie.Title.2023.1080P.BluRay.x264-GROUP", release.FHD},

		{"Abbreviated UHD", "Movie.Title.2023.4K.BluRay.x265-GROUP", release.UHD},
		{"HDR indicator", "Movie.Title.2023.HDR.2160p.WEB.x265-GROUP", release.UHD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := release.ParseResolution(tt.filename)
			assert.Equal(t, tt.expected, result, "Filename: %s", tt.filename)
		})
	}
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"German", "Movie.Title.2023.German.1080p.BluRay.x264-GROUP", "german"},
		{"French", "Movie.Title.2023.French.1080p.BluRay.x264-GROUP", "french"},
		{"Spanish", "Movie.Title.2023.Spanish.1080p.BluRay.x264-GROUP", "spanish"},
		{"Dutch", "Movie.Title.2023.Dutch.1080p.BluRay.x264-GROUP", "dutch"},
		{"Swedish", "Movie.Title.2023.Swedish.1080p.BluRay.x264-GROUP", "swedish"},
		{"Norwegian", "Movie.Title.2023.Norwegian.1080p.BluRay.x264-GROUP", "norwegian"},
		{"Finnish", "Movie.Title.2023.Finnish.1080p.BluRay.x264-GROUP", "finnish"},
		{"Danish", "Movie.Title.2023.Danish.1080p.BluRay.x264-GROUP", "danish"},
		{"Hebrew", "Movie.Title.2023.Hebrew.1080p.BluRay.x264-GROUP", "hebrew"},

		{"Capitalized German", "Movie.Title.2023.GERMAN.1080p.BluRay.x264-GROUP", "german"},
		{"Title Case French", "Movie.Title.2023.French.1080p.BluRay.x264-GROUP", "french"},

		{"German Subbed", "Movie.Title.2023.German.Subbed.1080p.BluRay.x264-GROUP", ""},
		{"No Language", "Movie.Title.2023.1080p.BluRay.x264-GROUP", ""},
		{"Subbed Only", "Movie.Title.2023.Subbed.1080p.BluRay.x264-GROUP", ""},
		{"Subbed German", "Movie.Title.2023.subbed.german.1080p.BluRay.x264-GROUP", ""},

		{"German in Title", "Der.German.Film.2023.1080p.BluRay.x264-GROUP", "german"},
		{"Spanish in Title", "El.Spanish.Film.2023.1080p.BluRay.x264-GROUP", "spanish"},

		{"Multi Languages", "Movie.Title.2023.MULTI.French.German.1080p.BluRay.x264-GROUP", "french"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := release.ParseLanguage(tt.filename)
			assert.Equal(t, tt.expected, result, "Filename: %s", tt.filename)
		})
	}
}
