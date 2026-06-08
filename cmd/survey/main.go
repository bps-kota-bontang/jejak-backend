package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"jejak/container"
	"jejak/internal/dto"

	"github.com/joho/godotenv"
)

const (
	modeCreate     = "create"
	modeSync       = "sync"
	modeSyncRegion = "sync-region"
	modeAnalyze    = "analyze"

	// Set mode di sini: modeCreate atau modeSync.
	runMode = modeSync

	// Set credential/payload di sini agar tidak perlu lewat arg command.
	fasihSurveyID       = "2561fe7e-c7b6-4e4a-b2d9-c8d254cba2bd"
	fasihSurveyPeriodID = "c7ee8024-e2fa-4d4a-b463-85f7ee508702"
	fasihXSRFToken      = "1262f43b-dfa8-4f41-b268-59f6e41e6aa9"
	fasihCookie         = "f5avraaaaaaaaaaaaaaaa_session_=OHBOOMNFBNGEELBOIOHECAOPGABMOPIKCIIFKGOJDPHCKCOLGHGKLGMHKCPOOFJNBIGDAMIOBKPNCKFOHGFAGMJPJAIPDPDAJDPLGNJCCPMPHEOFKMIAGAFPBAJGGJCI; f5avraaaaaaaaaaaaaaaa_session_=CJCPFKOGBHMKLDJCOPBGPFJIJKGPLMIGCDJGGPGCKAKLENOLHAGPKJGHFHPMOEFHEKIDPJDABKFNIABNIDPAFNDHJAOINEDIMCCMPDDHBBGAJNGEFIFDHBJMBCLJLNHP; f5_cspm=1234; f5avraaaaaaaaaaaaaaaa_session_=DJHEDPDLKPPPFBIILMNGALMBIHJECLGHIMOEMJCLFLCOKCHKDOEOKOCNMDMDFBGIEAIDJHBBCKLPDDKHMKNAGPBOJADODLAFMEDBPHLKHBHBJGOJKAIPPGAJFIKFJKKB; cf_clearance=aLhtNHSo9PiUQ81Sn2BXHB73L9TglXwyjKsTQanni24-1780467264-1.2.1.1-OZ4PwcuA3lqLxcavG.KGT1M5aH6F3f9_zt198NIxO78kOVxUDSj15z5mKfk1J3IYFxe287X9.VruNzdfrF5svRUH3huAUqBpstWF6NfztGLTB2F0wBlPAK3JlIUgZnazpwNJhhPCicPr8ikMaKhPPO.D54izAlN14F__QZsT7p.iCwz4cjusu7f2fAyb6T2ReYdSHuhGUGxBtlPCVUzzi.aWQ6NFrO_4ZFDOTWxSr_A0pc3RXfifZGM68NQUFMqFlcQyuUTpYzOsiUdGfP3FU_BIJsJB6nangxx0AYviQHHayKaW5KW5fDBZ2W2o02yCNN0J.SgTEangKdhFLjFZLg; _ga=GA1.3.454617234.1780467271; _ga_XXTTVXWHDB=GS2.3.s1780467270$o1$g1$t1780467325$j5$l0$h0; db8ca2b43ed851cc93e71fd5fd72bff7=6e9fe7597edb2bc3a2de73203448c109; XSRF-TOKEN=1262f43b-dfa8-4f41-b268-59f6e41e6aa9; SESSION=815b808d-a2a5-4259-adf5-7d8a70b34b6a; f5avraaaaaaaaaaaaaaaa_session_=GMIIFGFACIIEEEKEHFIEPIHIGKLMELAKIFKCEPNEAKJOCEEJBACGGKELJAGPKICHIEADPKECOMAOIDCJECNAHONBOMNCHIEADIAPEFMBAIOJFLAAMLAHONLINOHFJINE; TS01433fd3=01266d26d07eaf606970b22d8ecbbab451119108eb9ee40affd8d20c628376c9757f0ed347991a42568fc46a5ff5b2bdfdf1319b4c; TS01bafd94=01266d26d08b58d1150b1ffc3f0e5376d6fafd4c55fffa4f96d3d04013748073975769ba654a2c07e363b7b60fc53a8c9bf520f346; TS5d9b593f027=0868f8be6fab2000780fe123841c7c49ce161941ce98afa2a209649e2cf915f7e888e3d040fdc33008fa5337161130002d701eefcabd3c3d24237b8b7c6106070bcddbfc334c3bcc3b7a0e2d4af48854b5f88f03cf9676cedf824aa5981a84c6"

	syncAssignmentErrorStatus = -1
	syncAssignmentStatusAlias = "SUBMITTED BY Pencacah"
	syncFilterTargetType      = "TARGET_ONLY"

	// Optional: kosongkan agar otomatis ambil dari survey by id.
	syncRegionGroupID = ""
)

func main() {
	loadDotEnvIfPresent()

	surveyContainer, err := InitializeSurvey()
	if err != nil {
		log.Fatalf("failed to initialize survey command: %v", err)
	}

	mode := resolveRunMode()

	switch mode {
	case modeCreate:
		runCreate(surveyContainer)
	case modeSync:
		runSync(surveyContainer)
	case modeSyncRegion:
		runSyncRegion(surveyContainer)
	case modeAnalyze:
		runAnalyze(surveyContainer)
	default:
		log.Fatalf("invalid mode: %s (allowed: %s, %s, %s, %s)", mode, modeCreate, modeSync, modeSyncRegion, modeAnalyze)
	}
}

func resolveRunMode() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	return runMode
}

func loadDotEnvIfPresent() {
	if os.Getenv("APP_ENV") == "production" {
		return
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file not loaded: %v", err)
	}
}

func runCreate(surveyContainer *container.SurveyContainer) {
	if fasihSurveyID == "" || fasihSurveyPeriodID == "" || fasihXSRFToken == "" || fasihCookie == "" {
		log.Fatal("set fasihSurveyID, fasihSurveyPeriodID, fasihXSRFToken, and fasihCookie in code")
	}

	createReq := dto.CreateSurveyRequest{
		SurveyID:       fasihSurveyID,
		SurveyPeriodID: fasihSurveyPeriodID,
		XSRFToken:      fasihXSRFToken,
		Cookie:         fasihCookie,
	}

	if err := surveyContainer.SurveyService.CreateSurvey(createReq); err != nil {
		log.Fatalf("failed to create survey: %v", err)
	}

	fmt.Println("survey saved")
}

func runSync(surveyContainer *container.SurveyContainer) {
	if fasihSurveyPeriodID == "" {
		log.Fatal("set fasihSurveyPeriodID in code")
	}

	result, err := surveyContainer.SurveyService.SyncSurveyAssignments(context.Background(), fasihSurveyPeriodID)
	if err != nil {
		log.Fatalf("failed to sync survey: %v", err)
	}

	fmt.Printf("sync done: total=%d assignments=%d logs=%d answers=%d\n", result.TotalAssignments, result.SavedAssignments, result.SavedLogs, result.SavedAnswers)
}

func runAnalyze(surveyContainer *container.SurveyContainer) {
	if fasihSurveyPeriodID == "" {
		log.Fatal("set fasihSurveyPeriodID in code")
	}

	result, err := surveyContainer.SurveyService.AnalyzeSurvey(context.Background(), fasihSurveyPeriodID)
	if err != nil {
		log.Fatalf("failed to analyze survey: %v", err)
	}

	fmt.Printf("analyze done: surveyPeriodID=%s analyzed=%d/%d\n", result.SurveyPeriodID, result.AnalyzedAssignments, result.TotalAssignments)
	for _, assignment := range result.Assignments {
		fmt.Printf("- assignment=%s totalAnswers=%d locations=%d\n", assignment.AssignmentID, assignment.TotalAnswers, len(assignment.Locations))
	}
}

func runSyncRegion(surveyContainer *container.SurveyContainer) {
	if fasihSurveyPeriodID == "" {
		log.Fatal("set fasihSurveyPeriodID in code")
	}

	result, err := surveyContainer.SurveyService.SyncSurveyRegions(context.Background(), fasihSurveyPeriodID, dto.SyncSurveyRegionsRequest{
		RegionGroupID: syncRegionGroupID,
	})
	if err != nil {
		log.Fatalf("failed to sync survey regions: %v", err)
	}

	fmt.Printf("sync region done: regionGroupID=%s levelCount=%d saved=%d\n", result.RegionGroupID, result.LevelCount, result.SavedRegions)
}
