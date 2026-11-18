import Foundation
import LocalAuthentication

let ctx = LAContext()
var err: NSError?
guard ctx.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &err) else {
    FileHandle.standardError.write(Data("touch id unavailable\n".utf8))
    exit(2)
}
let sem = DispatchSemaphore(value: 0)
var ok = false
ctx.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, localizedReason: "Approve Intra sensitive grant") { success, _ in
    ok = success
    sem.signal()
}
sem.wait()
exit(ok ? 0 : 1)
